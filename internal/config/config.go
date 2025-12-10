package config

import (
	"os"
	"strconv"
)


type Config struct {
	// APP
	AppEnv string
	Port   string

	// Database
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	JWTSecret string
	ClickUpToken string
	ClickUpTeamID string
	ClickUpAPIKey string
	ClickUpSpaceID string
	ClickUpFolderID string
	ClickUpListID string

	// Admin login
	AdminUsername string
	AdminPassword string

	// Underload ≤ 35 hours/week
	// Normal 36–45 hours/week
	// Overload ≥ 60 hours/week
	WorkloadUnderload float64
	WorkloadNormalMin float64
	WorkloadNormalMax float64
	WorkloadOverload  float64
}

func Load() (*Config, error) {
	cfg := &Config{
		// App
		AppEnv: getEnv("APP_ENV", "development"),
		Port:   getEnv("PORT", "8001"),

		// DB
		DBHost: getEnv("DB_HOST", "db.fsufakerljrkzrlrjiwm.supabase.co"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASS", "aufa"),
		DBName: getEnv("DB_NAME", "kinerja_db"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "secret123"),

		// ClickUp
		ClickUpToken: getEnv("CLICKUP_TOKEN", "pk_101582122_8YV9NZHLPHQ75C9TWGM4RHB0U9MZJ2C2"),
		ClickUpTeamID: getEnv("CLICKUP_TEAM_ID", "90181837104"),
		ClickUpSpaceID: getEnv("CLICKUP_SPACE_ID", "90189431467"),
		ClickUpAPIKey: getEnv("CLICKUP_API_KEY", "pk_101582122_8YV9NZHLPHQ75C9TWGM4RHB0U9MZJ2C2"),
		ClickUpFolderID: getEnv("CLICKUP_FOLDER_ID", "90189201519"),
		ClickUpListID: getEnv("CLICKUP_LIST_ID", "901812499939"),

		// Admin login
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "dnakinerja-2025"),

		// Workload settings
		WorkloadUnderload: getEnvFloat("WORKLOAD_UNDERLOAD", 35),
		WorkloadNormalMin: getEnvFloat("WORKLOAD_NORMAL_MIN", 36),
		WorkloadNormalMax: getEnvFloat("WORKLOAD_NORMAL_MAX", 45),
		WorkloadOverload:  getEnvFloat("WORKLOAD_OVERLOAD", 60),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
