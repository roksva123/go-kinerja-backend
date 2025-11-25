package model

type ClickupTask struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    TextContent string `json:"text_content"`

    Status struct {
        ID     string `json:"id"`
        Status string `json:"status"`
        Type   string `json:"type"`
        Color  string `json:"color"`
    } `json:"status"`

    DateDone   string `json:"date_done"`
    DateClosed string `json:"date_closed"`
    DueDate    string `json:"due_date"`

    Assignees []struct {
        ID       int64  `json:"id"`
        IDString string `json:"id_string"`
        Username string `json:"username"`
        Email    string `json:"email"`
        Color    string `json:"color"`
    } `json:"assignees"`
}