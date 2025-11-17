package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type ClickUpService struct {
	Repo   *repository.PostgresRepo
	Token  string
	TeamID string
}

func NewClickUpService(repo *repository.PostgresRepo, token, teamID string) *ClickUpService {
	return &ClickUpService{Repo: repo, Token: token, TeamID: teamID}
}

func (s *ClickUpService) SyncTasks(ctx context.Context) error {
	page := 0
	for {
		req, _ := http.NewRequest("GET",
			fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/task?page=%d", s.TeamID, page),
			nil,
		)
		req.Header.Set("Authorization", s.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var data struct {
			Tasks []map[string]interface{} `json:"tasks"`
		}
		json.NewDecoder(resp.Body).Decode(&data)

		if len(data.Tasks) == 0 {
			break
		}

		for _, v := range data.Tasks {
			task := repository.Task{
				ID:   v["id"].(string),
				Name: v["name"].(string),
			}

			if assignees, ok := v["assignees"].([]interface{}); ok {
				if len(assignees) > 0 {
					task.EmployeeID = sqlNullStr(assignees[0].(map[string]interface{})["id"].(string))
				}
			}

			if status, ok := v["status"].(map[string]interface{}); ok {
				task.Status = sqlNullStr(status["status"].(string))
			}

			if timeEst, ok := v["time_estimate"].(float64); ok {
				task.TimeEstimateSeconds = sqlNullInt(int64(timeEst / 1000))
			}
			if timeSpent, ok := v["time_spent"].(float64); ok {
				task.TimeSpentSeconds = sqlNullInt(int64(timeSpent / 1000))
			}

			if sd, ok := v["start_date"].(string); ok && sd != "" {
				t, _ := time.Parse(time.RFC3339, sd)
				task.StartDate = sqlNullTime(t)
			}

			if dd, ok := v["due_date"].(string); ok && dd != "" {
				t, _ := time.Parse(time.RFC3339, dd)
				task.DueDate = sqlNullTime(t)
			}

			s.Repo.UpsertTask(ctx, task)
		}

		page++
	}

	return nil
}

func sqlNullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func sqlNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func sqlNullInt(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

