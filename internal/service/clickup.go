package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type ClickUpService struct {
	Repo   *repository.PostgresRepo
	Token  string
	TeamID string
	Client *http.Client
}

func NewClickUpService(repo *repository.PostgresRepo, token, teamID string) *ClickUpService {
	return &ClickUpService{
		Repo:   repo,
		Token:  token,
		TeamID: teamID,
		Client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *ClickUpService) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, method, url, nil)
	req.Header.Set("Authorization", s.Token)
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clickup api error %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (s *ClickUpService) SyncMembers(ctx context.Context) error {
	if s.TeamID == "" {
		return errors.New("team id not configured")
	}
	url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/member", s.TeamID)
	b, err := s.doRequest(ctx, "GET", url)
	if err != nil {
		return err
	}
	var out struct {
		Members []struct {
			User struct {
				ID         string `json:"id"`
				Email      string `json:"email"`
				Username   string `json:"username"`
				ProfilePic string `json:"profile_pic"`
			} `json:"user"`
		} `json:"members"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	for _, m := range out.Members {

		_, _ = s.Repo.UpsertEmployeeFromClickUp(ctx, m.User.Username, m.User.Email, m.User.ID)
	}
	return nil
}

func (s *ClickUpService) SyncTasks(ctx context.Context) (int, error) {
	if s.TeamID == "" {
		return 0, errors.New("team id not configured")
	}
	page := 0
	total := 0
	for {
		url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/task?page=%d", s.TeamID, page)
		b, err := s.doRequest(ctx, "GET", url)
		if err != nil {
			return total, err
		}
		var data struct {
			Tasks []map[string]interface{} `json:"tasks"`
		}
		if err := json.Unmarshal(b, &data); err != nil {
			return total, err
		}
		if len(data.Tasks) == 0 {
			break
		}
		for _, v := range data.Tasks {
			id, _ := v["id"].(string)
			name, _ := v["name"].(string)
			var employeeSQL sql.NullString
			if arr, ok := v["assignees"].([]interface{}); ok && len(arr) > 0 {
				if a0, ok2 := arr[0].(map[string]interface{}); ok2 {
					if aid, ok3 := a0["id"].(string); ok3 && aid != "" {
						if emp, err := s.Repo.GetEmployeeByClickUpID(ctx, aid); err == nil {
							employeeSQL = sql.NullString{String: emp.ID, Valid: true}
						} else {
							username := ""
							if un, ok4 := a0["username"].(string); ok4 {
								username = un
							}
							email := ""
							if em, ok5 := a0["email"].(string); ok5 {
								email = em
							}
							if empID, err2 := s.Repo.UpsertEmployeeFromClickUp(ctx, username, email, aid); err2 == nil {
								employeeSQL = sql.NullString{String: empID, Valid: true}
							}
						}
					}
				}
			}

			var statusSQL sql.NullString
			if st, ok := v["status"].(map[string]interface{}); ok {
				if sname, ok2 := st["status"].(string); ok2 && sname != "" {
					statusSQL = sql.NullString{String: sname, Valid: true}
				}
			}
			var timeEstSec sql.NullInt64
			if te, ok := v["time_estimate"].(float64); ok && te > 0 {
				timeEstSec = sql.NullInt64{Int64: int64(te) / 1000, Valid: true}
			} else if teStr, ok := v["time_estimate"].(string); ok && teStr != "" {
				if parsed, err := parseFloatStringToInt64(teStr); err == nil && parsed > 0 {
					timeEstSec = sql.NullInt64{Int64: parsed / 1000, Valid: true}
				}
			}

			var timeSpentSec sql.NullInt64
			if ts, ok := v["time_spent"].(float64); ok && ts > 0 {
				timeSpentSec = sql.NullInt64{Int64: int64(ts) / 1000, Valid: true}
			} else if tsStr, ok := v["time_spent"].(string); ok && tsStr != "" {
				if parsed, err := parseFloatStringToInt64(tsStr); err == nil && parsed > 0 {
					timeSpentSec = sql.NullInt64{Int64: parsed / 1000, Valid: true}
				}
			}

			// percent_complete
			var percentSQL sql.NullFloat64
			if pc, ok := v["percent_complete"].(float64); ok {
				percentSQL = sql.NullFloat64{Float64: pc, Valid: true}
			} else if pcStr, ok := v["percent_complete"].(string); ok && pcStr != "" {
				if pf, err := parseStringToFloat64(pcStr); err == nil {
					percentSQL = sql.NullFloat64{Float64: pf, Valid: true}
				}
			}
			var startSQL sql.NullTime
			if sd, ok := v["start_date"]; ok && sd != nil {
				if sdStr, ok := sd.(string); ok && sdStr != "" {
					if t, err := tryParseTime(sdStr); err == nil {
						startSQL = sql.NullTime{Time: t, Valid: true}
					}
				} else if sdNum, ok := sd.(float64); ok && sdNum > 0 {
					t := time.Unix(0, int64(sdNum)*int64(time.Millisecond))
					startSQL = sql.NullTime{Time: t, Valid: true}
				}
			}

			var dueSQL sql.NullTime
			if dd, ok := v["due_date"]; ok && dd != nil {
				if ddStr, ok := dd.(string); ok && ddStr != "" {
					if t, err := tryParseTime(ddStr); err == nil {
						dueSQL = sql.NullTime{Time: t, Valid: true}
					}
				} else if ddNum, ok := dd.(float64); ok && ddNum > 0 {
					t := time.Unix(0, int64(ddNum)*int64(time.Millisecond))
					dueSQL = sql.NullTime{Time: t, Valid: true}
				}
			}
			projectSQL := sql.NullString{}
			if pid, ok := v["project_id"].(string); ok && pid != "" {
				projectSQL = sql.NullString{String: pid, Valid: true}
			} else if pobj, ok := v["project"].(map[string]interface{}); ok {
				if pid2, ok2 := pobj["id"].(string); ok2 && pid2 != "" {
					projectSQL = sql.NullString{String: pid2, Valid: true}
				}
			}

			task := repository.Task{
				ID:                  id,
				Name:                name,
				EmployeeID:          employeeSQL,
				ProjectID:           projectSQL,
				Status:              statusSQL,
				TimeEstimateSeconds: timeEstSec,
				TimeSpentSeconds:    timeSpentSec,
				PercentComplete:     percentSQL,
				StartDate:           startSQL,
				DueDate:             dueSQL,
			}

			_ = s.Repo.UpsertTask(ctx, task)
			total++
		}
		page++
	}
	return total, nil
}

func tryParseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if parsed, err := parseFloatStringToInt64(s); err == nil && parsed > 0 {
		return time.Unix(0, parsed*int64(time.Millisecond)), nil
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", s)
}

func parseFloatStringToInt64(s string) (int64, error) {
	var f float64
	_, err := fmt.Sscan(s, &f)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

func parseStringToFloat64(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscan(s, &f)
	return f, err
}
