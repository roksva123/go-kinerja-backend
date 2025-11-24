package model

import "time"

type User struct {
    ID         int64 `json:"id"`         
    ClickUpID  int64  `json:"clickup_id"` 
    Username   string `json:"username"`
    Name       string `json:"name"`
    Email      string `json:"email"`
    Role       string `json:"role"`
    Color      string `json:"color"`
    PasswordHash string `json:"-"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}



type ClickUpUser struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Color    string `json:"color"`
    Role     string `json:"role"`
}