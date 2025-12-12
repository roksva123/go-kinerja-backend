package handlers

type TaskInResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	StatusID          string  `json:"status_id"`
	StatusName        string  `json:"status_name"`
	StatusType        string  `json:"status_type"`
	DateDone          *string `json:"date_done,omitempty"`
	ProjectName       *string `json:"project_name,omitempty"`
	DateClosed        *string `json:"date_closed,omitempty"`
	StartDate         *string `json:"start_date,omitempty"`
	DueDate           *string `json:"due_date,omitempty"`
	TimeEstimateHours float64 `json:"time_estimate_hours"`
	TimeSpentHours    float64 `json:"time_spent_hours"`
	Category          string  `json:"category"`
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