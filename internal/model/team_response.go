package model

type TeamResponse struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    ParentID string `json:"parent_id"`
}
