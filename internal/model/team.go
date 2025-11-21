package model

import "time"

type Team struct {
    ID        string    `json:"id"`        
    Name      string    `json:"name"`
    ParentID  *string   `json:"parent_id"` 
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

