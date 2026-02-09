package config_env

import (
	"fmt"
	"os"
)

type AppEnv struct {
	Port           string
	Seed           bool
	DatabaseConfig DbConfig
}

type DbConfig struct {
	DSN string
}

func LoadAppEnv() (*AppEnv, error) {
	db_dsn := os.Getenv("DATABASE_DSN")
	appPort := os.Getenv("APP_PORT")
	seed := os.Getenv("SEED") == "true"

	if appPort == "" {
		appPort = "8080"
	}
	if db_dsn == "" {
		return nil, fmt.Errorf("DATABASE_DSN")
	}
	return &AppEnv{
		Port: appPort,
		Seed: seed,
		DatabaseConfig: DbConfig{
			DSN: db_dsn,
		},
	}, nil
}
