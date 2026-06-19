package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL   string
	KafkaBrokers  string
	RedisURL      string
	GRPCPort      string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://crypto2krw:crypto2krw_dev@localhost:5432/crypto2krw?sslmode=disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9093"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		GRPCPort:     getEnv("GRPC_PORT", "50051"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
