package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

// DB holds all repository implementations
type DB struct {
	pool *pgxpool.Pool
	// Repository implementations
	Users   users.Querier
	Shopify shopify.Querier
}

// New creates a new database connection pool using DATABASE_URL from config
func New() (*DB, error) {
	return NewWithURL(config.Values.Database.URL)
}

// RunMigrations runs database migrations with out-of-order support
func RunMigrations(databaseURL string) error {
	// Create a standard sql.DB connection for goose
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to open database for migrations")
	}
	defer sqlDB.Close()

	// Set migration dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return errors.Wrap(err, "failed to set goose dialect")
	}

	// Enable out-of-order migrations
	goose.SetSequential(false)

	// Get the migrations directory path
	migrationsDir := filepath.Join("migrations")

	// Run migrations
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	return nil
}

// NewWithURL creates a new database connection pool with the provided URL
func NewWithURL(databaseURL string) (*DB, error) {
	// Parse and configure the connection pool
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse database URL")
	}

	// Configure pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	// Create the connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create connection pool")
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, "failed to ping database")
	}

	// Initialize repositories
	userQueries := users.New(pool)
	shopifyQueries := shopify.New(pool)

	return &DB{
		pool:    pool,
		Users:   userQueries,
		Shopify: shopifyQueries,
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.pool.Close()
}

// Pool returns the underlying pgxpool.Pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// WithTx executes a function within a database transaction
func (db *DB) WithTx(ctx context.Context, fn func(*TxDB) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	txDB := &TxDB{
		Users:   users.New(tx),
		Shopify: shopify.New(tx),
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(txDB); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Wrapf(err, "transaction failed, rollback also failed: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// TxDB provides repository access within a transaction
type TxDB struct {
	Users   users.Querier
	Shopify shopify.Querier
}
