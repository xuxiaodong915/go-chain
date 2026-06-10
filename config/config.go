package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort string
	DBPath     string // SQLite file path, used when MYSQL_DSN is not set
	DSN        string // MySQL DSN, e.g. user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
}

func Load() *Config {
	dsn := os.Getenv("MYSQL_DSN")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/recipes.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else if port[0] != ':' {
		port = ":" + port
	}

	return &Config{
		ServerPort: port,
		DBPath:     dbPath,
		DSN:        dsn,
	}
}

// DailyRecommendExpiry defines how often daily recommendation rotates
const DailyRecommendExpiry = 24 * time.Hour
