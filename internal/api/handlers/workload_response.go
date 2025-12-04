package handlers

// TaskInResponse adalah struct untuk tugas yang akan ditampilkan di dalam respons,
// tanpa menyertakan field 'assignees' yang berulang.
type TaskInResponse struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	TextContent       string   `json:"text_content,omitempty"`
	StatusName        string   `json:"status_name"`
	StartDate         *string  `json:"start_date"`
	DueDate           *string  `json:"due_date"`
	DateDone          *string  `json:"date_done,omitempty"`
	TimeEstimateHours float64  `json:"time_estimate_hours"`
	TimeSpentHours    float64  `json:"time_spent_hours"`
}

// AssigneeWithTasks adalah struct untuk menampung data assignee beserta daftar tugasnya.
type AssigneeWithTasks struct {
	ClickupID int                `json:"clickup_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Name      string             `json:"name"`
	Tasks     []TaskInResponse   `json:"tasks"`
}

// TasksByAssigneeResponse adalah struct untuk respons akhir dari endpoint.
type TasksByAssigneeResponse struct {
	Count     int                 `json:"count"`
	Assignees []AssigneeWithTasks `json:"assignees"`
}