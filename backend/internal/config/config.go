package config

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var (
	Values  *serviceConfig
	Version string
)

const envKey = "GO_ENV"

func loadConfigFile(path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}

	if err := godotenv.Load(path); err != nil {
		panic(errors.Wrapf(err, "failed to load env file: %s", path))
	}
}

func init() {
	env := os.Getenv(envKey)
	if env == "" {
		env = "development"
	}
	os.Setenv(envKey, env)

	currDir, err := os.Getwd()
	if err != nil {
		panic(errors.Wrapf(err, "error getting pwd"))
	}

	loadConfigFile(filepath.Join(currDir, ".env."+env+".local"))
	loadConfigFile(filepath.Join(currDir, ".env.local"))
	loadConfigFile(filepath.Join(currDir, ".env."+env))
	loadConfigFile(filepath.Join(currDir, ".env"))

	cfg, err := parseConfig()
	if err != nil {
		log.Printf("error parsing config: %s", err.Error())
		panic(errors.Wrapf(err, "error parsing config"))
	}

	if err := validateConfig(cfg); err != nil {
		panic(errors.Wrapf(err, "invalid configuration"))
	}

	Values = cfg
}

func IsTest() bool {
	return Values.Env == "test" || flag.Lookup("test.v") != nil
}

func IsTestOrDevelopment() bool {
	return strings.ToLower(Values.Service.Env) != "production" && strings.ToLower(Values.Service.Env) != "staging"
}

func IsProduction() bool {
	return strings.ToLower(Values.Service.Env) == "production"
}

func parseConfig() (*serviceConfig, error) {
	var cfg serviceConfig
	if _, err := flags.NewParser(&cfg, flags.Default|flags.IgnoreUnknown).Parse(); err != nil {
		return nil, errors.Wrapf(err, "parsing flags")
	}

	return &cfg, nil
}

func validateConfig(cfg *serviceConfig) error {
	// Validate encryption secret key
	if len(cfg.Encryption.SecretKey) != 32 {
		return errors.Errorf("ENCRYPTION_SECRET_KEY must be exactly 32 bytes, got %d", len(cfg.Encryption.SecretKey))
	}

	return nil
}
