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
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "127.0.0.1"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASS", "aufa"),
		getEnv("DB_NAME", "kinerja_db"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	username := getEnv("ADMIN_USERNAME", "admin")
	password := getEnv("ADMIN_PASSWORD", "dnakinerja-2025")

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	db.Exec("DELETE FROM admins")
	_, err = db.Exec("INSERT INTO admins (username, password_hash) VALUES ($1, $2)", username, string(hash))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin created:", username)
}

func getEnv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
