package model

type Task struct {
    ID          string `db:"id" json:"id"`
    Name        string `db:"name" json:"name"`
    Status      string `db:"status" json:"status"`
    AssigneeID  int64  `db:"assignee_id" json:"assignee_id"`
    DueDate     int64  `db:"due_date" json:"due_date"`
}
