package model

import (
	"time"
)

type User struct {
    ClickUpID    int64     `json:"clickup_id"`
    RoleID       JsonNullInt64 `json:"role_id,omitempty"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    Role         string    `json:"role"`
    Status       string    `json:"status"`
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Color    string `json:"color"`
}

// type ClickupTask struct {
//     ID          string `json:"id"`
//     Name        string `json:"name"`
//     Description string `json:"description"`
//     TextContent string `json:"text_content"`

//     Status struct {
//         ID     string `json:"id"`
//         Status string `json:"status"`
//         Type   string `json:"type"`
//         Color  string `json:"color"`
//     } `json:"status"`

//     DateDone   string `json:"date_done"`
//     DateClosed string `json:"date_closed"`
//     DueDate    string `json:"due_date"`

//     Assignees []struct {
//         ID       int64  `json:"id"`
//         IDString string `json:"id_string"`
//         Username string `json:"username"`
//         Email    string `json:"email"`
//         Color    string `json:"color"`
//     } `json:"assignees"`
// }
type ClickupTag struct {
	Name string `json:"name"`
}

type ClickupPriority struct {
	ID       string `json:"id"`
	Priority string `json:"priority"`
	Color    string `json:"color"`
}

type ClickupTaskResponse struct {
	Tasks []ClickupTask `json:"tasks"`
}
