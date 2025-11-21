package model

import "time"

type Admin struct {
	ID        string    `json:"id"` 
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Employee struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    TeamID    string    `json:"team_id"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
	Color     string `json:"color"`
}
