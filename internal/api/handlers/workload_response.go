package handlers

type TaskInResponse struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	TextContent       string   `json:"text_content,omitempty"`
	ProjectName       *string  `json:"project_name,omitempty"`
	StatusName        string   `json:"status_name"`
	StartDate         *string  `json:"start_date"`
	DueDate           *string  `json:"due_date"`
	DateDone          *string  `json:"date_done,omitempty"`
	TimeEstimateHours float64  `json:"time_estimate_hours"`
	TimeSpentHours    float64  `json:"time_spent_hours"`
}

type AssigneeWithTasks struct {
	ClickupID       int              `json:"clickup_id"`
	Username        string           `json:"username"`
	Email           string           `json:"email"`
	Name            string           `json:"name"`
	TotalSpentHours float64          `json:"total_spent_hours"`
	ExpectedHours   float64          `json:"expected_hours"`
	TotalTasks      int              `json:"total_tasks"`
	TotalWorkHours  float64          `json:"total_work_hours"`
	ActualWorkHours float64          `json:"actual_work_hours"`
	TotalUpcomingHours float64       `json:"total_upcoming_hours"`
	Tasks           []TaskInResponse `json:"tasks"`
}

type TasksByAssigneeResponse struct {
	Count     int                 `json:"count"`
	Assignees []AssigneeWithTasks `json:"assignees"`
}