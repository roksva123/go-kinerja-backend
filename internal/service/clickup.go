package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

// ClickUpService handles fetching/syncing data from ClickUp API.
type ClickUpService struct {
	Repo   *repository.PostgresRepo
	Token  string
	TeamID string
	Client *http.Client
}

// NewClickUpService creates a new service instance.
func NewClickUpService(repo *repository.PostgresRepo, token, teamID string) *ClickUpService {
	return &ClickUpService{
		Repo:   repo,
		Token:  token,
		TeamID: teamID,
		Client: &http.Client{Timeout: 20 * time.Second},
	}
}

// doRequest does an authenticated GET to ClickUp and returns body bytes.
func (s *ClickUpService) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
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

// SyncMembers fetches team members and upserts to employees table.
func (s *ClickUpService) SyncMembers(ctx context.Context) error {
	if strings.TrimSpace(s.TeamID) == "" {
		return errors.New("clickup team id not configured")
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
				CustomID   string `json:"custom_id"`
				// more fields ignored
			} `json:"user"`
		} `json:"members"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	for _, m := range out.Members {
		username := m.User.Username
		email := m.User.Email
		clickupID := m.User.ID
		// Upsert into employees table
		_, _ = s.Repo.UpsertEmployeeFromClickUp(ctx, username, email, clickupID)
	}
	return nil
}

// SyncTasks fetches tasks from ClickUp (paginated) and upserts to local tasks table.
// It returns number of tasks processed.
func (s *ClickUpService) SyncTasks(ctx context.Context) (int, error) {
	if strings.TrimSpace(s.TeamID) == "" {
		return 0, errors.New("clickup team id not configured")
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
			// id & name
			id, _ := v["id"].(string)
			name, _ := v["name"].(string)

			// Assignee -> internal employee id (first assignee)
			var employeeSQL sql.NullString
			if arr, ok := v["assignees"].([]interface{}); ok && len(arr) > 0 {
				if a0, ok2 := arr[0].(map[string]interface{}); ok2 {
					if aid, ok3 := a0["id"].(string); ok3 && aid != "" {
						// try find employee by clickup id
						if emp, err := s.Repo.GetEmployeeByClickUpID(ctx, aid); err == nil && emp != nil {
							employeeSQL = sql.NullString{String: emp.ID, Valid: true}
						} else {
							// upsert minimal employee
							username := ""
							if un, ok4 := a0["username"].(string); ok4 {
								username = un
							}
							email := ""
							if em, ok5 := a0["email"].(string); ok5 {
								email = em
							}
							if empID, err2 := s.Repo.UpsertEmployeeFromClickUp(ctx, username, email, aid); err2 == nil && empID != "" {
								employeeSQL = sql.NullString{String: empID, Valid: true}
							}
						}
					}
				}
			}

			// status
			var statusSQL sql.NullString
			if st, ok := v["status"].(map[string]interface{}); ok {
				if sname, ok2 := st["status"].(string); ok2 && sname != "" {
					statusSQL = sql.NullString{String: sname, Valid: true}
				}
			}

			// time_estimate and time_spent (ms)
			var timeEstSec sql.NullInt64
			if te, ok := tryFloatFromInterface(v["time_estimate"]); ok && te > 0 {
				timeEstSec = sql.NullInt64{Int64: int64(te) / 1000, Valid: true}
			}
			var timeSpentSec sql.NullInt64
			if ts, ok := tryFloatFromInterface(v["time_spent"]); ok && ts > 0 {
				timeSpentSec = sql.NullInt64{Int64: int64(ts) / 1000, Valid: true}
			}

			// percent_complete
			var percentSQL sql.NullFloat64
			if pc, ok := tryFloat64FromInterface(v["percent_complete"]); ok {
				percentSQL = sql.NullFloat64{Float64: pc, Valid: true}
			}

			// start_date and due_date: string or number (ms)
			var startSQL sql.NullTime
			if sd, ok := v["start_date"]; ok && sd != nil {
				if t, err := tryParseTime(sd); err == nil {
					startSQL = sql.NullTime{Time: t, Valid: true}
				}
			}
			var dueSQL sql.NullTime
			if dd, ok := v["due_date"]; ok && dd != nil {
				if t, err := tryParseTime(dd); err == nil {
					dueSQL = sql.NullTime{Time: t, Valid: true}
				}
			}

			// project id: sometimes at "project_id" or nested "project"
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

// helper: try parse float from interface (handles float64, string numeric, json.Number)
func tryFloatFromInterface(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return f, true
		}
	case string:
		// sometimes ClickUp returns numeric as string
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func tryFloat64FromInterface(v interface{}) (float64, bool) {
	return tryFloatFromInterface(v)
}

// tryParseTime: accepts RFC3339 string, or numeric milliseconds (float64/int/string/json.Number)
func tryParseTime(v interface{}) (time.Time, error) {
	if v == nil {
		return time.Time{}, errors.New("nil")
	}
	switch t := v.(type) {
	case string:
		// try RFC3339
		if t == "" {
			return time.Time{}, errors.New("empty")
		}
		if tt, err := time.Parse(time.RFC3339, t); err == nil {
			return tt, nil
		}
		// numeric string ms
		if parsed, err := strconv.ParseInt(t, 10, 64); err == nil && parsed > 0 {
			return time.Unix(0, parsed*int64(time.Millisecond)), nil
		}
		return time.Time{}, fmt.Errorf("unsupported time string: %s", t)
	case float64:
		if t <= 0 {
			return time.Time{}, errors.New("invalid")
		}
		return time.Unix(0, int64(t)*int64(time.Millisecond)), nil
	case int64:
		if t <= 0 {
			return time.Time{}, errors.New("invalid")
		}
		return time.Unix(0, t*int64(time.Millisecond)), nil
	case json.Number:
		if parsed, err := t.Int64(); err == nil && parsed > 0 {
			return time.Unix(0, parsed*int64(time.Millisecond)), nil
		}
		if f, err := t.Float64(); err == nil && f > 0 {
			return time.Unix(0, int64(f)*int64(time.Millisecond)), nil
		}
	}
	return time.Time{}, errors.New("unsupported")
}

