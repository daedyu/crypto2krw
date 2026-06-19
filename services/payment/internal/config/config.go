package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	RedisURL        string
	JWTAccessSecret string
	OracleURL       string
	CoreInternalURL string
}

func Load() (*Config, error) {
	secret := getEnv("JWT_ACCESS_SECRET", "")
	if secret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	return &Config{
		Port:            getEnv("PORT", "3003"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTAccessSecret: secret,
		OracleURL:       getEnv("ORACLE_URL", "http://localhost:3002"),
		CoreInternalURL: getEnv("CORE_INTERNAL_URL", "http://localhost:8080"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
