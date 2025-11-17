package config

import "os"

type Config struct {
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPass        string
	DBName        string
	JWTSecret     string
	ClickUpToken  string
	ClickUpTeamID string
}

func LoadFromEnv() *Config {
	return &Config{
		Port:          os.Getenv("PORT"),
		DBHost:        getenv("DB_HOST", "localhost"),
		DBPort:        getenv("DB_PORT", "5432"),
		DBUser:        getenv("DB_USER", "postgres"),
		DBPass:        getenv("DB_PASS", "aufa"),
		DBName:        getenv("DB_NAME", "kinerja_db"),
		JWTSecret:     getenv("JWT_SECRET", "replace_this_jwt_secret"),
		ClickUpToken:  os.Getenv("CLICKUP_TOKEN"),
		ClickUpTeamID: os.Getenv("CLICKUP_TEAM_ID"),
	}
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
