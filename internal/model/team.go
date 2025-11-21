package model

type Team struct {
    TeamID   string `db:"team_id" json:"team_id"`
    Name     string `db:"name" json:"name"`
    ParentID string `db:"parent_id" json:"parent_id"`
}
