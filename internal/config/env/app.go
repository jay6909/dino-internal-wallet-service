package config_env

import (
	"fmt"
	"os"
)

type AppEnv struct {
	DatabaseConfig DbConfig
}

type DbConfig struct {
	DSN string
}

func LoadAppEnv() (*AppEnv, error) {
	db_dsn := os.Getenv("DATABASE_DSN")

	if db_dsn == "" {
		return nil, fmt.Errorf("DATABASE_DSN")
	}
	return &AppEnv{
		DatabaseConfig: DbConfig{
			DSN: db_dsn,
		},
	}, nil
}
