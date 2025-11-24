package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Build DSN from env
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "127.0.0.1"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASS", "aufa"),
		getEnv("DB_NAME", "kinerja_db"),
	)

	// Connect database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed connect DB:", err)
	}
	defer db.Close()

	// Test DB connection
	if err := db.Ping(); err != nil {
		log.Fatal("DB unreachable:", err)
	}

	// Ensure table "admins" exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS admins (
			id SERIAL PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("Failed ensure admins table:", err)
	}

	// Read env (fallback if not provided)
	username := getEnv("ADMIN_USERNAME", "admin")
	password := getEnv("ADMIN_PASSWORD", "dnakinerja-2025")

	// Hash password
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Clear old admin (optional)
	_, err = db.Exec("DELETE FROM admins")
	if err != nil {
		log.Fatal("Failed delete old admins:", err)
	}

	// Insert new admin
	_, err = db.Exec(
		"INSERT INTO admins (username, password_hash) VALUES ($1, $2)",
		username, string(hash),
	)
	if err != nil {
		log.Fatal("Failed insert admin:", err)
	}

	fmt.Println("Admin created successfully!")
	fmt.Println("Username:", username)
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
