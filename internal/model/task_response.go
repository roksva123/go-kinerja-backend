package model

type TaskResponse struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    TextContent string `json:"text_content"`
    Description string `json:"description"`

    Status struct {
        ID    string `json:"id"`
        Name  string `json:"name"`
        Type  string `json:"type"`
        Color string `json:"color"`
    } `json:"status"`

    DateCreated *int64 `json:"date_created"`
    DateDone    *int64 `json:"date_done"`
    DateClosed  *int64 `json:"date_closed"`

    AssignedTo  string `json:"assigned_to"`
    Username string `json:"display_name"`
    Email    string `json:"email"`
    Color    string `json:"color"`
}
