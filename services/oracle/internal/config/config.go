package config

import "os"

type Config struct {
	Port     string
	RedisURL string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "3002"),
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
