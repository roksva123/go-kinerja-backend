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

	TimeEfficiencyPercentage *float64 `json:"time_efficiency_percentage,omitempty"`
	RemainingTimeHours       *float64 `json:"-"`
	RemainingTimeFormatted   *string  `json:"remaining_time,omitempty"`
	ActualDurationFormatted  *string  `json:"actual_duration,omitempty"`
	ScheduleStatus           *string  `json:"schedule_status,omitempty"`
}

type AssigneeWithTasks struct {
	ClickupID                  int64            `json:"clickup_id"`
	Username                   string           `json:"username"`
	Email                      string           `json:"email"`
	Name                       string           `json:"name"`
	TotalSpentHours            float64          `json:"total_spent_hours"`
	ExpectedHours              float64          `json:"expected_hours"`
	TotalTasks                 int              `json:"total_tasks"`
	ActualWorkHours            float64          `json:"actual_work_hours"`
	TotalUpcomingHours         float64          `json:"total_upcoming_hours"`
	OnTimeCompletionPercentage *float64         `json:"on_time_completion_percentage,omitempty"`
	Tasks                      []TaskInResponse `json:"tasks"`
}

type TasksByAssigneeResponse struct {
	Count     int                 `json:"count"`
	Assignees []AssigneeWithTasks `json:"assignees"`
}