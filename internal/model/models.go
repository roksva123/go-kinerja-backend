package model

import "time"

type Admin struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Employee struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Color     string `json:"color"`
}
