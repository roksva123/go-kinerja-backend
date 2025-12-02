package model

import "time"

type Task struct {
	ID          string `json:"id"`
	TaskID      string `json:"task_id"`
	Name        string `json:"name"`
	TextContent string `json:"text_content"`
	Description string `json:"description"`
	Status      struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Type  string `json:"type"`
		Color string `json:"color"`
	} `json:"status"`
	DateDone   *int64 `json:"date_done,omitempty"`
	DateClosed *int64 `json:"date_closed,omitempty"`

	Username string `json:"username"`
	Email    string `json:"email"`
	Color    string `json:"color"`

	TimeEstimateMs *int64 `json:"time_estimate_ms,omitempty"`
	TimeSpentMs    *int64 `json:"time_spent_ms,omitempty"`

	StartDate    *int64 `json:"start_date"`
	DueDate      *int64 `json:"due_date"`
	DateCreated  *int64 `json:"date_created"`
	TimeEstimate *int64 `json:"time_estimate"`

	AssigneeUserID    *int64 `json:"assignee_user_id"`
	AssigneeClickUpID *int64 `json:"assignee_id"`
	AssigneeColor     string  `json:"assingeescolor"`
	OwnerUsername     string `json:"owner_username"`
	OwnerEmail        string `json:"owner_email"`
	AssigneeUsername  string `json:"assignee_username"`
	AssigneeEmail     string `json:"assignee_email"`
}

type TaskWithMember struct {
	TaskID          string `json:"task_id"`
	TaskName        string `json:"task_name"`
	TaskDescription string `json:"description"`
	TaskStatus      string `json:"status"`
	TaskStatusType  string `json:"status_type"`
	TaskStatusColor string `json:"status_color"`

	StartDate    *int64 `json:"start_date"`
	DueDate      *int64 `json:"due_date"`
	DateCreated  *int64 `json:"date_created"`
	DateDone     *int64 `json:"date_done"`
	DateClosed   *int64 `json:"date_closed"`
	TimeEstimate *int64 `json:"time_estimate"`
	TimeSpent    *int64 `json:"time_spent"`

	UserID   int64  `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	Color    string `json:"color,omitempty"`

	TeamID   *string `json:"team_id,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
}

type TaskFull struct {
	TaskID       string `json:"task_id"`
	TaskName     string `json:"task_name"`
	Description  string `json:"description"`
	StatusName   string `json:"status_name"`
	StatusType   string `json:"status_type"`
	StatusColor  string `json:"status_color"`
	StartDate    *int64 `json:"start_date"`
	DueDate      *int64 `json:"due_date"`
	DateDone     *int64 `json:"date_done"`
	DateClosed   *int64 `json:"date_closed"`
	TimeEstimate *int64 `json:"time_estimate"`
	TimeSpent    *int64 `json:"time_spent"`

	// Member
	UserID   *int64  `json:"user_id"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Color    *string `json:"color"`
	Role     *string `json:"role"`

	// Team
	TeamID   *string `json:"team_id"`
	TeamName *string `json:"team_name"`
}

type TaskItem struct {
	ID          int64  `json:"id"`
	TaskID      string `json:"task_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TextContent string `json:"text_content"`

	StatusID    int64  `json:"status_id"`
	StatusName  string `json:"status_name"`
	StatusType  string `json:"status_type"`
	StatusColor string `json:"status_color"`

	DateDone   *time.Time `json:"date_done"`
	DateClosed *time.Time `json:"date_closed"`

	TimeSpent int64  `json:"time_spent"`
	Category  string `json:"category"`
}


type TaskResponse struct {
    ID          string  `json:"id"`
    TaskID      string  `json:"task_id"`
	CustomID    string  `json:"costom"`
	CustomItemID  int64 `json:"custom_item_id"`
    Name        string  `json:"name"`
    TextContent string  `json:"text_content"`
    Description string  `json:"description"`
	Category string `json:"category"`

    Status struct {
        ID    string `json:"id"`
        Name  string `json:"name"`
        Type  string `json:"type"`
        Color string `json:"color"`
    } `json:"status"`

    DateDone   *int64 `json:"date_done,omitempty"`
    DateClosed *int64 `json:"date_closed,omitempty"`

    Username string `json:"username"`
    Email    string `json:"email"`
    Color    string `json:"color"`

    TimeEstimateMs *int64 `json:"time_estimate_ms,omitempty"`
    TimeSpentMs    *int64 `json:"time_spent_ms,omitempty"`

    StartDate    *int64 `json:"start_date"`
    DueDate      *int64 `json:"due_date"`
    DateCreated  *int64 `json:"date_created"`
	DateUpdated  *int64 `json:"date_updated"`
    TimeEstimate *int64 `json:"time_estimate"`
	Assignees     []TaskAssignee  `json:"assignees"`

	AssigneeClickUpID *string `json:"assignee_clickup_id"`
    AssigneeUserID   *int64 `json:"assignee_user_id"`
    AssigneeID       *int64 `json:"assignee_id"`
	AssigneeColor    *string `json:"assignee_color"`
    AssigneeUsername *string `json:"assignee_username"`
    AssigneeEmail    *string `json:"assignee_email"`
}



type Status struct {
    ID        string `json:"id"`
    Name      string `json:"status"`
    Color     string `json:"color"`
    Type      string `json:"type"`
    OrderIndex int    `json:"orderindex"`
}


type TaskAssignee struct {
    ID        int64   `json:"id"`
    Username  string  `json:"username"`
    Color     string  `json:"color"`
    Initials  string  `json:"initials"`
    Email     string  `json:"email"`
    Profile   *string `json:"profilePicture"`
}
