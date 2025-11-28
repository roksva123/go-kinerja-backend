package model


type WorkloadUser struct {
    UserID        int64           `json:"user_id"`
    Username      string          `json:"username"`
    Email         string          `json:"email"`
    Role          string          `json:"role"`
    Color         string          `json:"color"`
    TotalHours    float64         `json:"total_hours"`
    TaskCount     int             `json:"task_count"`
    TotalTasks    int64           `json:"total_tasks"`
    TotalMs       int64           `json:"total_ms"`
    Tasks         []TaskResponse  `json:"tasks"`
    ByStatus      map[string]float64 `json:"by_status"`
    ByCategory    map[string]float64 `json:"by_category"`
    StandardHours float64         `json:"standard_hours"`
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