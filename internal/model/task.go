package model

import "database/sql"

type Task struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	EmployeeID          sql.NullInt64   `json:"employee_id"`
	ProjectID           sql.NullString  `json:"project_id"`
	Status              sql.NullString  `json:"status"`
	TimeEstimateSeconds sql.NullInt64   `json:"time_estimate_seconds"`
	TimeSpentSeconds    sql.NullInt64   `json:"time_spent_seconds"`
	PercentComplete     sql.NullFloat64 `json:"percent_complete"`
	StartDate           sql.NullTime    `json:"start_date"`
	DueDate             sql.NullTime    `json:"due_date"`
	UpdatedAt           sql.NullTime    `json:"updated_at"`
}