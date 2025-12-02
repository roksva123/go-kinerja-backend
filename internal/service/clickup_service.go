package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type ClickUpService struct {
    Repo   *repository.PostgresRepo
    APIKey string
    Token  string
    TeamID string
    Client *http.Client
    DB     *sql.DB
}

func NewClickUpService(
    repo *repository.PostgresRepo,
    apiKey string,
    token string,
    teamID string,
    db *sql.DB,
) *ClickUpService {

    return &ClickUpService{
        Repo:   repo,
        APIKey: apiKey,
        Token:  token,
        TeamID: teamID,
        Client: &http.Client{Timeout: 20 * time.Second},
        DB:     db,
    }
}


func (s *ClickUpService) doRequest(ctx context.Context, method, url string) ([]byte, error) {
    req, _ := http.NewRequestWithContext(ctx, method, url, nil)
    req.Header.Set("Authorization", s.Token)
    res, err := s.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()
    body, _ := io.ReadAll(res.Body)
    if res.StatusCode >= 400 {
        return nil, fmt.Errorf("clickup api error %d: %s", res.StatusCode, string(body))
    }
    return body, nil
}

// SyncTeam 
func (s *ClickUpService) SyncTeam(ctx context.Context) error {
    if s.TeamID == "" {
        return errors.New("team id not configured")
    }
    url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/space", s.TeamID)
    b, err := s.doRequest(ctx, "GET", url)
    if err != nil {
        return err
    }
    var out struct {
        Spaces []struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        } `json:"spaces"`
    }
    if err := json.Unmarshal(b, &out); err != nil {
        return err
    }
    if len(out.Spaces) == 0 {
        return errors.New("no spaces found for this team")
    }
    for _, sp := range out.Spaces {
        if err := s.Repo.UpsertTeam(ctx, sp.ID, sp.Name, ""); err != nil {
            return err
        }
    }
    return nil
}

// SyncMembers 
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
				ID       int64  `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Color    string `json:"color"`
			} `json:"user"`
			Role int `json:"role"`
		} `json:"members"`
	}

	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}

	for _, member := range out.Members {
		// Mapping ke model
		u := &model.User{
			ID:          member.User.ID,
			ClickUpID:   member.User.ID,
			DisplayName: member.User.Username,
			Name:        member.User.Username, // Default name to username
			Email:       member.User.Email,
			Role:        "member", // Anda bisa menambahkan logika untuk role di sini
			Color:       member.User.Color,
		}

		if err := s.Repo.UpsertUser(ctx, u); err != nil {
			fmt.Println("ERROR UPSERT USER:", err)
			return err
		}
	}

	return nil
}

func parseInt64Ptr(v interface{}) *int64 {
    switch val := v.(type) {
    case float64:
        x := int64(val)
        return &x
    case string:
        if val == "" { return nil }
        if x, err := strconv.ParseInt(val, 10, 64); err == nil {
            return &x
        }
    }
    return nil
}

func anyToInt64Ptr(v interface{}) *int64 {
    return parseInt64Ptr(v)
}


// SyncTasks
func (s *ClickUpService) SyncTasks(ctx context.Context) (int, error) {
    if s.TeamID == "" {
        return 0, errors.New("team id not configured")
    }

    log.Println("=== START SYNC TASKS ===")

    page := 0
    total := 0

    for {
        url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/task?page=%d", s.TeamID, page)
        log.Println("[REQUEST]", url)

        b, err := s.doRequest(ctx, "GET", url)
        if err != nil {
            log.Println("❌ REQUEST ERROR:", err)
            return total, err
        }

        var out struct {
            Tasks []map[string]interface{} `json:"tasks"`
        }

        if err := json.Unmarshal(b, &out); err != nil {
            log.Println("❌ JSON PARSE ERROR:", err)
            return total, err
        }

        log.Printf("[PAGE %d] FOUND %d TASKS\n", page, len(out.Tasks))

        if len(out.Tasks) == 0 {
            log.Println("=== NO MORE TASKS — FINISHED ===")
            break
        }

        for _, raw := range out.Tasks {

            rawJSON, _ := json.Marshal(raw)
            log.Printf("--- RAW TASK DATA FROM CLICKUP ---\n%s\n---------------------------------", string(rawJSON))


            t := &model.TaskResponse{}
            log.Println("---- PROCESSING TASK ----")

            // ID
            if id, ok := raw["id"].(string); ok {
                t.ID = id
            }
            log.Println("ID:", t.ID)

            // Name
            if name, ok := raw["name"].(string); ok {
                t.Name = name
            }

            // Content
            if txt, ok := raw["text_content"].(string); ok {
                t.TextContent = txt
            }
            if desc, ok := raw["description"].(string); ok {
                t.Description = desc
            }

            // Status
            if st, ok := raw["status"].(map[string]interface{}); ok {
                if v, ok := st["id"].(string); ok { t.Status.ID = v }
                if v, ok := st["status"].(string); ok { t.Status.Name = v }
                if v, ok := st["type"].(string); ok { t.Status.Type = v }
                if v, ok := st["color"].(string); ok { t.Status.Color = v }
            }
            log.Printf("Status: %+v\n", t.Status)

            // Dates
            t.DateDone = parseInt64Ptr(raw["date_done"])
            t.DateClosed = parseInt64Ptr(raw["date_closed"])
            t.DateCreated = parseInt64Ptr(raw["date_created"])
            t.DateUpdated = parseInt64Ptr(raw["date_updated"])
            t.StartDate = parseInt64Ptr(raw["start_date"])
            t.DueDate = parseInt64Ptr(raw["due_date"])

            if t.StartDate == nil {
                t.StartDate = t.DateCreated
                log.Printf("StartDate is nil, using DateCreated (%v) instead.", t.DateCreated)
            }
            if t.DueDate == nil {
                t.DueDate = t.StartDate 
                if t.DateDone != nil {
                    t.DueDate = t.DateDone
                }
                log.Printf("DueDate is nil, using a fallback date (%v).", t.DueDate)
            }

            // Time Estimate
            t.TimeEstimate = parseInt64Ptr(raw["time_estimate"])
            t.TimeSpentMs = parseInt64Ptr(raw["time_spent"])

            // Custom Fields
            if cfArr, ok := raw["custom_fields"].([]interface{}); ok {
                for _, rawCF := range cfArr {
                    cf := rawCF.(map[string]interface{})

                    name, _ := cf["name"].(string)
                    val := cf["value"]
                    if val == nil { continue }

                    // Tanggal Mulai
                    if name == "Tanggal Mulai" {
                        t.StartDate = anyToInt64Ptr(val)
                    }

                    // Tanggal Akhir
                    if name == "Tanggal Akhir" {
                        t.DueDate = anyToInt64Ptr(val)
                    }
                }
            }

            // Assignee
            if arr, ok := raw["assignees"].([]interface{}); ok && len(arr) > 0 {
                if a0, ok := arr[0].(map[string]interface{}); ok {

                    if uname, ok := a0["username"].(string); ok {
                        t.AssigneeUsername = &uname
                    }
                    if email, ok := a0["email"].(string); ok {
                        t.AssigneeEmail = &email
                    }
                    if col, ok := a0["color"].(string); ok {
                        t.AssigneeColor = &col
                    }
                    if id, ok := a0["id"].(float64); ok {
                        v := fmt.Sprintf("%.0f", id)
                        t.AssigneeClickUpID = &v
						uid := int64(id)
						t.AssigneeUserID = &uid
                    }
                }
            }

            log.Printf("Assignee: %v (%v)\n", t.AssigneeUsername, t.AssigneeEmail)

            // UPSERT
            log.Println("UPSERT TASK:", t.ID)
            if err := s.Repo.UpsertTask(ctx, t); err != nil {
                log.Println("❌ UPSERT ERROR:", err)
                return total, err
            }

            log.Println("✔ UPSERT SUCCESS:", t.ID)
            total++
        }

        page++
    }

    log.Println("=== SYNC COMPLETE — TOTAL:", total)
    return total, nil
}


func ptrString(s string) *string {
    return &s
}
// func ptrInt64(v int64) *int64 {
//     return &v
// }

func (s *ClickUpService) FullSync(ctx context.Context) ([]model.FullSync, error) {

    members, err := s.Repo.GetMembers(ctx)
    if err != nil {
        return nil, err
    }

    tasks, err := s.Repo.GetTasks(ctx)
    if err != nil {
        return nil, err
    }

    var out []model.FullSync

    for _, t := range tasks {
        var matchedMember *model.User

        for _, m := range members {
        if m.DisplayName == t.Username || m.Email == t.Email {
            matchedMember = &m
            break
            }
        }


        fs := model.FullSync{
            TaskID:      t.ID,
            TaskName:    t.Name,
            TaskStatus:  t.Status.Name,
            DateCreated: t.DateClosed,
            DateDone:    t.DateDone,
            AssignedTo:  t.Username,
        }

        if matchedMember != nil {
            fs.UserID = matchedMember.ID
            fs.DisplayName = matchedMember.DisplayName
            fs.Email = matchedMember.Email
            fs.Role = matchedMember.Role
            fs.Color = matchedMember.Color
        }

        out = append(out, fs)
    }

    return out, nil
}


func (s *ClickUpService) GetTasks(ctx context.Context) ([]model.TaskResponse, error) {
    return s.Repo.GetTasks(ctx)
}

func (s *ClickUpService) GetMembers(ctx context.Context) ([]model.User, error) {
    return s.Repo.GetMembers(ctx)
}

func (s *ClickUpService) GetTeams(ctx context.Context) ([]model.Team, error) {
    return s.Repo.GetTeams(ctx)
}

func (s *ClickUpService) FullSyncFiltered(ctx context.Context, filter model.FullSyncFilter) ([]model.FullSync, error) {
    now := time.Now()

    if filter.Range == "last_6_months" {
        end := now.UnixMilli()
        start := now.AddDate(0, -6, 0).UnixMilli()
        filter.StartDate = &start
        filter.EndDate = &end
    }

    if filter.Range == "next_6_months" {
        start := now.UnixMilli()
        end := now.AddDate(0, 6, 0).UnixMilli()
        filter.StartDate = &start
        filter.EndDate = &end
    }

    // Ambil dari repo
    data, err := s.Repo.GetFullSyncFiltered(ctx, filter.StartDate, filter.EndDate, filter.Role)
    if err != nil {
        return nil, err
    }

    if len(data) == 0 {
        return nil, errors.New("Tidak ada task pada rentang tanggal yang dipilih")
    }

    var out []model.FullSync

    for _, t := range data {

        // convert milliseconds -> jam
        convert := func(ms *int64) float64 {
            if ms == nil {
                return 0
            }
            return float64(*ms) / 1000 / 60 / 60
        }

        fs := model.FullSync{
            TaskID:     t.TaskID,
            TaskName:   t.TaskName,
            TaskStatus: t.TaskStatus,

            DateCreated: t.DateCreated,
            DateDone:    t.DateDone,
            DateClosed:  t.DateClosed,

            HoursCreated: convert(t.DateCreated),
            HoursDone:    convert(t.DateDone),
            HoursClosed:  convert(t.DateClosed),

            UserID:   t.UserID,
            DisplayName: t.Username,
            Email:    t.Email,
            Role:     t.Role,
            Color:    t.Color,
        }

        out = append(out, fs)
    }

    return out, nil
}


func (s *ClickUpService) PullTasks(ctx context.Context) (int, error) {
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

        var out struct {
            Tasks []map[string]interface{} `json:"tasks"`
        }
        if err := json.Unmarshal(b, &out); err != nil {
            return total, err
        }
        if len(out.Tasks) == 0 { break }

        for _, raw := range out.Tasks {
            t := &model.TaskResponse{}

            if v, ok := raw["id"].(string); ok { t.ID = v }
            if v, ok := raw["name"].(string); ok { t.Name = v }
            if v, ok := raw["text_content"].(string); ok { t.TextContent = v }
            if v, ok := raw["description"].(string); ok { t.Description = v }

            if st, ok := raw["status"].(map[string]interface{}); ok {
                t.Status.ID = safeString(st["id"])
                t.Status.Name = safeString(st["name"])
                t.Status.Type = safeString(st["type"])
                t.Status.Color = safeString(st["color"])
                t.DateCreated = toIntPtr(raw["date_created"])
            }

            // helper to parse int64 from either string/float
            toIntPtr := func(x interface{}) *int64 {
                if x == nil { return nil }
                switch v := x.(type) {
                case string:
                    if v == "" { return nil }
                    if n, err := strconv.ParseInt(v, 10, 64); err == nil { return &n }
                case float64:
                    n := int64(v); return &n
                case int64:
                    n := v; return &n
                }
                return nil
            }

            t.DateDone = toIntPtr(raw["date_done"])
            t.DateClosed = toIntPtr(raw["date_closed"])
            t.StartDate = toIntPtr(raw["start_date"])
            t.DueDate = toIntPtr(raw["due_date"])
            t.TimeEstimate = toIntPtr(raw["time_estimate"])

            // assignee first
            if arr, ok := raw["assignees"].([]interface{}); ok && len(arr) > 0 {
                if a, ok := arr[0].(map[string]interface{}); ok {
                    t.Username = safeString(a["username"])
                    t.Email = safeString(a["email"])
                    t.Color = safeString(a["color"])
                    // if clickup id present:
                    if cid, ok := a["id"].(float64); ok {
                        v := int64(cid)
                        // you can add an AssigneeClickUpID field in model if wanted
                        _ = v
                    }
                }
            }

            if err := s.Repo.UpsertTask(ctx, t); err != nil {
                return total, err
            }
            total++
        }

        page++
    }

    return total, nil
}

func safeString(v any) string {
    switch val := v.(type) {
    case nil:
        return ""
    case string:
        return val
    default:
        return fmt.Sprintf("%v", val)
    }
}

func (s *WorkloadService) SyncAll(ctx context.Context) error {
    return nil
}


func (s *ClickUpService) FullSyncFlow(ctx context.Context, filter model.FullSyncFilter) ([]model.TaskWithMember, error) {
    if _, err := s.PullTasks(ctx); err != nil {
        return nil, err
    }
    return s.Repo.GetFullDataFiltered(ctx, filter.StartDate, filter.EndDate, filter.Role, filter.Username)
}

func toIntPtr(v interface{}) *int64 {
	switch val := v.(type) {
	case float64:
		n := int64(val)
		return &n
	case int64:
		return &val
	case int:
		n := int64(val)
		return &n
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return &i
		}
	}
	return nil
}

func (s *ClickUpService) AllSync(ctx context.Context) error {

    // 1. Sync Team
    if err := s.SyncTeam(ctx); err != nil {
        return err
    }

    // 2. Sync Members
    if err := s.SyncMembers(ctx); err != nil {
        return err
    }

    // 3. Sync Tasks
    _, err := s.SyncTasks(ctx)
    if err != nil {
        return err
    }

    return nil
}

func normalizeStatus(status string) string {
	lowerStatus := strings.ToLower(status)
	if strings.Contains(lowerStatus, "review") || strings.Contains(lowerStatus, "progress") {
		return "progres"
	}
	if strings.Contains(lowerStatus, "do") { // "to do"
		return "to do"
	}
	if strings.Contains(lowerStatus, "done") || strings.Contains(lowerStatus, "complete") || strings.Contains(lowerStatus, "closed") {
		return "done"
	}
	if strings.Contains(lowerStatus, "cancel") {
		return "canceled"
	}
	return lowerStatus // return as is if no match
}

func (s *ClickUpService) GetWorkload(ctx context.Context, startMs, endMs int64) ([]model.WorkloadUser, error) {
	query := `
		SELECT
			u.id,
			u.name,
			u.username,
			u.email,
			u.role,
			u.color,
			t.id,
			t.name,
			t.status_name,
			t.start_date,
			t.due_date,
			t.date_done,
			t.time_spent_ms
			ua.name AS assignee_full_name -- Tambahkan ini untuk nama assignee
		FROM users u
		LEFT JOIN tasks t ON u.id = t.assignee_user_id AND (
			(t.start_date <= $2 AND t.due_date >= $1) OR
			(t.date_done >= $1 AND t.date_done <= $2)
		)
		ORDER BY u.name ASC, t.start_date ASC
	`
	rows, err := s.DB.QueryContext(ctx, query, startMs, endMs)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userMap := make(map[int64]*model.WorkloadUser)

	for rows.Next() {
		var userID int64
		var name, username, email, role, color string
		var taskID, taskName, statusName, assigneeFullName sql.NullString 
		var startDate, dueDate, dateDone, timeSpentMs sql.NullInt64

		err := rows.Scan(
			&userID, &name, &username, &email, &role, &color,
			&taskID, &taskName, &statusName,
			&startDate, &dueDate, &dateDone, &timeSpentMs, 
			&assigneeFullName,
		)
		if err != nil {
			return nil, err
		}

		if _, ok := userMap[userID]; !ok {
			userMap[userID] = &model.WorkloadUser{
				UserID:        userID,
				Name:          name,
				Username:      username,
				Email:         email,
				Role:          role,
				Color:         color,
				Tasks:         []model.TaskDetail{},
				ByStatus:      make(map[string]float64),
				ByCategory:    make(map[string]float64),
				StandardHours: 8.0,
			}
		}

		if taskID.Valid {
			task := model.TaskDetail{
				ID:         taskID.String,
				Name:       taskName.String,
				StatusName: normalizeStatus(statusName.String),
				StartDate:  msToDateString(toInt64Ptr(startDate)),
				DueDate:    msToDateString(toInt64Ptr(dueDate)),
				DateDone:   msToDateString(toInt64Ptr(dateDone)),
			}
			if assigneeFullName.Valid { 
				task.AssigneeName = &assigneeFullName.String
			}

			if timeSpentMs.Valid {
				hours := float64(timeSpentMs.Int64) / 3600000.0
				task.TimeSpentHours = hours
				userMap[userID].TotalHours += hours
			}

			userMap[userID].Tasks = append(userMap[userID].Tasks, task)
		}
	}

	var result []model.WorkloadUser
	for _, user := range userMap {
		user.TaskCount = len(user.Tasks)
		result = append(result, *user)
	}

	return result, nil
}

func msToDateString(ms *int64) *string {
	if ms == nil {
		return nil
	}
	t := time.UnixMilli(*ms)
	s := t.Format("2006-01-02")
	return &s
}

func (s *ClickUpService) GetTasksByRange(ctx context.Context, startMs, endMs int64) ([]model.TaskDetail, error) {
	query := `
		SELECT
			t.id,
			t.name,
			t.status_name,
			t.start_date,
			t.due_date,
			t.date_done,
			t.time_spent_ms,
			t.assignee_user_id,
			u.name
		FROM tasks t
		LEFT JOIN users u ON t.assignee_user_id = u.id
		WHERE
			-- Mencakup tugas yang aktif dalam rentang waktu
			(t.start_date <= $2 AND t.due_date >= $1)
			-- Mencakup tugas yang selesai dalam rentang waktu
			OR (t.date_done >= $1 AND t.date_done <= $2)
		ORDER BY t.start_date ASC
	`
	rows, err := s.DB.QueryContext(ctx, query, startMs, endMs)

	if err != nil {
		log.Printf("Error querying tasks by range: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []model.TaskDetail
	log.Println("--- Scanning Tasks By Range ---")
	for rows.Next() {
		var taskID, taskName, statusName string
		var startDate, dueDate, dateDone, timeSpentMs, assigneeUserID sql.NullInt64
		var assigneeName sql.NullString

		err := rows.Scan(
			&taskID,
			&taskName,
			&statusName,
			&startDate,
			&dueDate,
			&dateDone,
			&timeSpentMs,
			&assigneeUserID,
			&assigneeName,
		)
		if err != nil {
			log.Printf("Error scanning task row: %v", err)
			return nil, err
		}

		taskDetail := model.TaskDetail{
			ID:         taskID,
			Name:       taskName,
			StatusName: normalizeStatus(statusName),
			StartDate:  msToDateString(toInt64Ptr(startDate)),
			DueDate:    msToDateString(toInt64Ptr(dueDate)),
			DateDone:   msToDateString(toInt64Ptr(dateDone)),
		}

		if assigneeUserID.Valid {
			taskDetail.AssigneeUserID = &assigneeUserID.Int64
		}
		if assigneeName.Valid {
			taskDetail.AssigneeName = &assigneeName.String
		}
		if timeSpentMs.Valid {
			taskDetail.TimeSpentHours = float64(timeSpentMs.Int64) / 3600000.0
		}

		tasks = append(tasks, taskDetail)
	}
	log.Printf("--- Found a total of %d tasks in range ---", len(tasks))

	return tasks, nil
}

func toInt64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}