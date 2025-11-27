package model

import "time"

type Task struct {
    ID         string     `json:"id"`
    Name       string     `json:"name"`
    UserID        int64   `json:"user_id,omitempty"`
    Username      string  `json:"username,omitempty"`
    Email         string  `json:"email,omitempty"`
    Role          string  `json:"role,omitempty"`
    Color         string  `json:"color,omitempty"`
    	Status      struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
	} `json:"status"`
    AssigneeID int64      `json:"assignee_id"`
    StartDate     *int64  `json:"start_date"`
    DueDate       *int64  `json:"due_date"`
    DateCreated *int64 `json:"date_created"`
    DateDone      *int64  `json:"date_done"`
    DateClosed    *int64  `json:"date_closed"`
    TimeEstimate  *int64  `json:"time_estimate"`
    TimeSpent     *int64  `json:"time_spent"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    TaskID        string  `json:"task_id"`
    TaskName      string  `json:"task_name"`
    TaskStatusType string `json:"status_type"`
    TaskStatusColor string `json:"status_color"`
    TeamID        *string `json:"team_id,omitempty"`
    TeamName      *string `json:"team_name,omitempty"`
    TextContent string  `json:"text_content"`
    TimeEstimateMs *int64 `json:"time_estimate_ms,omitempty"`
	TimeSpentMs    *int64 `json:"time_spent_ms,omitempty"`
    Description string  `json:"description"`
}

type TaskWithMember struct {
    TaskID        string  `json:"task_id"`
    TaskName      string  `json:"task_name"`
    TaskDescription string `json:"description"`
    TaskStatus    string  `json:"status"`
    TaskStatusType string `json:"status_type"`
    TaskStatusColor string `json:"status_color"`

    StartDate     *int64  `json:"start_date"`
    DueDate       *int64  `json:"due_date"`
    DateCreated *int64 `json:"date_created"`
    DateDone      *int64  `json:"date_done"`
    DateClosed    *int64  `json:"date_closed"`
    TimeEstimate  *int64  `json:"time_estimate"`
    TimeSpent     *int64  `json:"time_spent"`

    UserID        int64   `json:"user_id,omitempty"`
    Username      string  `json:"username,omitempty"`
    Email         string  `json:"email,omitempty"`
    Role          string  `json:"role,omitempty"`
    Color         string  `json:"color,omitempty"`

    TeamID        *string `json:"team_id,omitempty"`
    TeamName      *string `json:"team_name,omitempty"`
}

type TaskFull struct {
    TaskID           string   `json:"task_id"`
    TaskName         string   `json:"task_name"`
    Description      string   `json:"description"`
    StatusName       string   `json:"status_name"`
    StatusType       string   `json:"status_type"`
    StatusColor      string   `json:"status_color"`
    StartDate        *int64   `json:"start_date"`
    DueDate          *int64   `json:"due_date"`
    DateDone         *int64   `json:"date_done"`
    DateClosed       *int64   `json:"date_closed"`
    TimeEstimate     *int64   `json:"time_estimate"`
    TimeSpent        *int64   `json:"time_spent"`

    // Member
    UserID           *int64   `json:"user_id"`
    Username         *string  `json:"username"`
    Email            *string  `json:"email"`
    Color            *string  `json:"color"`
    Role             *string  `json:"role"`

    // Team
    TeamID           *string  `json:"team_id"`
    TeamName         *string  `json:"team_name"`
}
