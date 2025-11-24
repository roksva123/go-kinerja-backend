package model

import "time"

type Member struct {
    ID          string    `json:"id"`               
    ClickUpID   int64     `json:"clickup_id"`       
    Username    string    `json:"username"`
    Email       string    `json:"email"`
    Color       string    `json:"color"`
    Photo       string    `json:"photo"`
    TeamID      string    `json:"team_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
