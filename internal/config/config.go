package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	WorkerCount int    `env:"WORKER_COUNT" envDefault:"5"`
	QueueSize   int    `env:"QUEUE_SIZE" envDefault:"100"`
}

func Load() (*Config, error) {
	db_url := os.Getenv("DATABASE_URL")
	if db_url == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	return &Config{
		DatabaseURL: db_url,
		WorkerCount: getInt("WORKER_COUNT", 5),
		QueueSize:   getInt("QUEUE_SIZE", 100),
	}, nil
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
