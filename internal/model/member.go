package model

import "time"

type Member struct {
    ID        int64     `json:"id"`
    ClickUpID string    `json:"clickup_id"`
    Username  string    `json:"username"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Color     string    `json:"color"`
    TeamID    string    `json:"team_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
