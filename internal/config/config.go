package config

import (
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	KafkaBrokers  []string
	PostgresDSN   string
	HTTPPort      string
	ConsumerGroup string
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	cfg := &Config{
		KafkaBrokers:  parseBrokers(getEnv("KAFKA_BROKERS", "localhost:19092")),
		PostgresDSN:   getEnv("POSTGRES_DSN", "postgres://streamlens:streamlens@localhost:5432/streamlens?sslmode=disable"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		ConsumerGroup: getEnv("CONSUMER_GROUP", "metrics-processor-group"),
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseBrokers(brokers string) []string {
	return strings.Split(brokers, ",")
}
