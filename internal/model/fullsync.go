
package model

type FullSync struct {
    TaskID       string  `json:"task_id"`
    TaskName     string  `json:"task_name"`
    TaskStatus   string  `json:"task_status"`

    DateCreated  *int64 `json:"date_created"`
    DateDone     *int64 `json:"date_done"`
    DateClosed   *int64 `json:"date_closed"`

    HoursCreated float64 `json:"hours_created"`
    HoursDone    float64 `json:"hours_done"`
    HoursClosed  float64 `json:"hours_closed"`

    UserID      int64  `json:"user_id"`
    DisplayName string `json:"display_name"`  
    Email       string `json:"email"`
    Role        string `json:"role"`
    Color       string `json:"color"`

    AssignedTo  string `json:"assigned_to"`
}



type FullSyncFilter struct {
    Username  string `json:"username"`
    Email     string `json:"email"`
    Status    string `json:"status"`
    StartDate *int64 `json:"start_date"`
    EndDate   *int64 `json:"end_date"`
    Range     string `json:"range"` 
    Role      string `json:"role"`
}

// type TaskWithMember struct {
//     TaskID      string
//     TaskName    string
//     TaskStatus  string
//     DateCreated *int64
//     DateDone    *int64
//     DateClosed  *int64

//     UserID   int64
//     Username string
//     Email    string
//     Role     string
//     Color    string
// }
