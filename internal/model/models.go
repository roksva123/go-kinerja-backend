package model

import "time"

type Admin struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Employee struct {
	ID         string     `json:"id"`
	Fullname   string     `json:"fullname"`
	Email      string     `json:"email,omitempty"`
	Position   string     `json:"position,omitempty"`
	ClickUpID  string     `json:"clickup_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Task struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	EmployeeID          string     `json:"employee_id,omitempty"`
	ProjectID           string     `json:"project_id,omitempty"`
	Status              string     `json:"status,omitempty"`
	TimeEstimateSeconds int64      `json:"time_estimate_seconds,omitempty"`
	TimeSpentSeconds    int64      `json:"time_spent_seconds,omitempty"`
	PercentComplete     float64    `json:"percent_complete,omitempty"`
	StartDate           *time.Time `json:"start_date,omitempty"`
	DueDate             *time.Time `json:"due_date,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}
