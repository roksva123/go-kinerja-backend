package utils

import "github.com/roksva123/go-kinerja-backend/internal/model"

func ConvertTeamToResponse(team model.Team) model.TeamResponse {
    return model.TeamResponse{
        ID:       team.TeamID,
        Name:     team.Name,
        ParentID: team.ParentID,
    }
}

func ConvertTeamsToResponse(teams []model.Team) []model.TeamResponse {
    resp := make([]model.TeamResponse, 0, len(teams))
    for _, t := range teams {
        resp = append(resp, ConvertTeamToResponse(t))
    }
    return resp
}
