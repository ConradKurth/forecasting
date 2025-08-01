package db

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// DB holds all repository implementations
type DB struct {
	pool *pgxpool.Pool
	// Repository implementations
	Users users.Querier
}

// New creates a new database connection pool using DATABASE_URL from config
func New() (*DB, error) {
	return NewWithURL(config.Values.Database.URL)
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

	return &DB{
		pool:  pool,
		Users: userQueries,
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
		Users: users.New(tx),
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
	Users users.Querier
}
