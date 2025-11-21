package utils

import "github.com/roksva123/go-kinerja-backend/internal/model"

func ConvertTeamToResponse(t model.Team) model.TeamResponse {
    return model.TeamResponse{
        ID:       t.ID,
        Name:     t.Name,
        ParentID: t.ParentID,
    }
}

func ConvertTeamsToResponse(items []model.Team) []model.TeamResponse {
    out := make([]model.TeamResponse, 0, len(items))
    for _, t := range items {
        out = append(out, ConvertTeamToResponse(t))
    }
    return out
}
