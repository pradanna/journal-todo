package config

import (
	"os"
)

type Config struct {
	AppPort          string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTAccessSecret  string
	JWTRefreshSecret string
	GoogleClientID   string
}

func Load() Config {
	return Config{
		AppPort:          getEnv("APP_PORT", "8081"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "jurnal"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
		DBName:           getEnv("DB_NAME", "jurnal_freelance"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		GoogleClientID:   getEnv("GOOGLE_CLIENT_ID", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
