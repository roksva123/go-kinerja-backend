package model

import "database/sql"

type WorkloadUser struct {
    UserID        int64           `json:"user_id"`
    Name          string          `json:"name"`
    Username      string          `json:"username"`
    Email         string          `json:"email"`
    Role          string          `json:"role"`
    Color         string          `json:"color"`
    TotalHours    float64         `json:"total_hours"`
    TaskCount     int             `json:"task_count"`
    TotalTasks    int64           `json:"-"` 
    TotalMs       int64           `json:"-"`
    Tasks         []TaskDetail    `json:"tasks"`
    ByStatus      map[string]float64 `json:"by_status"`
    ByCategory    map[string]float64 `json:"by_category"`
    StandardHours float64         `json:"standard_hours"`
}

type AssigneeDetail struct {
	ClickUpID int64  `json:"clickup_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

type TaskDetail struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	StatusName      string   `json:"status_name"`
	StartDate       *string  `json:"start_date"`
	DueDate         *string  `json:"due_date"`
	DateDone        *string  `json:"date_done,omitempty"`
	TimeSpentHours  float64  `json:"time_spent_hours"`
	Assignees       []AssigneeDetail `json:"assignees"`
}

type WorkloadSummary struct {
    TotalUsers int     `json:"total_users"`
    TotalHours float64 `json:"total_hours"`
    AvgHours   float64 `json:"avg_hours"`
}


type WorkloadResponse struct {
	Start       string          `json:"start"`
	End         string          `json:"end"`
	StandardHoursPerPerson float64 `json:"standard_hours_per_person"`
	Summary     struct {
		TotalUsers int     `json:"total_users"`
		TotalHours float64 `json:"total_hours"`
		AvgHours   float64 `json:"avg_hours"`
	} `json:"summary"`
	Users                 []WorkloadUser    `json:"users"`
}
type WorkloadFilter struct {
	Start int64
	End   int64
}

type WorkloadItem struct {
    ID               string        `json:"id"`
    TaskID           string        `json:"task_id"`
    Name             string        `json:"name"`
    TextContent      string        `json:"text_content"`
    Description      string        `json:"description"`
    StatusID         string        `json:"status_id"`
    StatusName       string        `json:"status_name"`
    StatusType       string        `json:"status_type"`
    StatusColor      string        `json:"status_color"`

    DateDone         sql.NullInt64 `json:"date_done"`
    DateClosed       sql.NullInt64 `json:"date_closed"`

    Username         string         `json:"username"`
    Email            string         `json:"email"`

    AssigneeUserID    sql.NullInt64 `json:"assignee_user_id"`
    AssigneeClickupID sql.NullInt64 `json:"assignee_clickup_id"`
    AssigneeUsername  string        `json:"assignee_username"`
    AssigneeEmail     string        `json:"assignee_email"`

    Color             string        `json:"color"`

    TimeEstimateMs    sql.NullInt64 `json:"time_estimate_ms"`
    TimeSpentMs       sql.NullInt64 `json:"time_spent_ms"`

    StartDate         sql.NullInt64 `json:"start_date"`
    DueDate           sql.NullInt64 `json:"due_date"`

    TimeEstimate      sql.NullInt64 `json:"time_estimate"`
}


type WorkloadRangeRequest struct {
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
    UserID    string `json:"user_id"` 
}
