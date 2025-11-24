package model


type TaskResponse struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    TextContent string  `json:"text_content"`
    Description string  `json:"description"`
    Status      struct {
        ID    string `json:"id"`
        Name  string `json:"name"`
        Type  string `json:"type"`
        Color string `json:"color"`
    } `json:"status"`
    DateDone   *int64  `json:"date_done"`
    DateClosed *int64  `json:"date_closed"`
    Username   string  `json:"username"`
    Email      string  `json:"email"`
    Color      string  `json:"color"`
}