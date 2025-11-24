package model

type Employee struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    TeamID    string    `json:"team_id"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
	Color     string `json:"color"`
}
