package model

import (
	"time"
)

type User struct {
  ID        int64     `json:"id"`
  Username  string    `json:"username"`
  Name      string    `json:"name"`
  Password  string    `json:"-"`
  Email     string    `json:"email"`
  Role      string    `json:"role"`
  Color     string    `json:"color"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

type ClickUpUser struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Color    string `json:"color"`
    Role     string `json:"role"`
}