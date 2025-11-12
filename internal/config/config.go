package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	PostgresDSN      string
	DBMaxConns       int32
	DBMinConns       int32
	DBMaxConnLifetime time.Duration
	DBHealthCheckPeriod time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	RunMigrations    bool
}

func Load() (*Config, error) {
	port := getEnvOrDefault("PORT", "8080")

	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		return nil, fmt.Errorf("POSTGRES_DSN обязателен")
	}

	dbMaxConns := getEnvAsInt32("DB_MAX_CONNS", 20)
	dbMinConns := getEnvAsInt32("DB_MIN_CONNS", 5)
	dbMaxConnLifetime := getEnvAsDuration("DB_MAX_CONN_LIFETIME_MS", 30*60*1000)
	dbHealthCheckPeriod := getEnvAsDuration("DB_HEALTH_CHECK_PERIOD_MS", 60*1000)
	readTimeout := getEnvAsDuration("READ_TIMEOUT_MS", 5000)
	writeTimeout := getEnvAsDuration("WRITE_TIMEOUT_MS", 10000)
	runMigrations := getEnvAsBool("RUN_MIGRATIONS", true)

	return &Config{
		Port:                port,
		PostgresDSN:         postgresDSN,
		DBMaxConns:          dbMaxConns,
		DBMinConns:          dbMinConns,
		DBMaxConnLifetime:   dbMaxConnLifetime,
		DBHealthCheckPeriod: dbHealthCheckPeriod,
		ReadTimeout:         readTimeout,
		WriteTimeout:        writeTimeout,
		RunMigrations:       runMigrations,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt32(key string, defaultValue int32) int32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(value)
}

func getEnvAsDuration(key string, defaultMs int64) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return time.Duration(defaultMs) * time.Millisecond
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return time.Duration(defaultMs) * time.Millisecond
	}
	return time.Duration(value) * time.Millisecond
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
