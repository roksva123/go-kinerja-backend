package model

import "time"

type Task struct {
    ID         string     `json:"id"`
    Name       string     `json:"name"`
    Status     string     `json:"status"`
    AssigneeID int64      `json:"assignee_id"`
    DueDate    int64      `json:"due_date"`
    DateDone   *int64     `json:"date_done"`
    DateClosed *int64     `json:"date_closed"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
}



