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

	"github.com/lib/pq"
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
	url := "https://api.clickup.com/api/v2/team"

	b, err := s.doRequest(ctx, "GET", url)
	if err != nil {
		return err
	}

	var out struct {
		Teams []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Members []struct {
				User struct {
					ID       int64  `json:"id"`
					Username string `json:"username"`
					Email    string `json:"email"`
					Color    string `json:"color"`
				} `json:"user"`
				Role int `json:"role"`
			} `json:"members"`
		} `json:"teams"`
	}

	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}

	for _, team := range out.Teams {
		if team.ID != s.TeamID {
			continue 
		}

		for _, member := range team.Members {
			u := &model.User{
				ClickUpID:   member.User.ID,
				Name: member.User.Username,
				Email:       member.User.Email,
				Status:      "aktif", 
				Role:        "backend", 
			}
			if member.Role == 2 {
				u.Role = "pm" 
			} else if member.Role == 4 {
				u.Role = "frontend"
			}

			if err := s.Repo.UpsertUser(ctx, u); err != nil {
				fmt.Println("ERROR UPSERT USER:", err)
				return err
			}
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

// Helper untuk konversi milidetik (int64) ke *time.Time
func msToTimePtr(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}
	t := time.UnixMilli(*ms)
	return &t
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
            t.DateDone = msToTimePtr(parseInt64Ptr(raw["date_done"]))
            t.DateClosed = msToTimePtr(parseInt64Ptr(raw["date_closed"]))
            t.DateCreated = msToTimePtr(parseInt64Ptr(raw["date_created"]))
            t.DateUpdated = msToTimePtr(parseInt64Ptr(raw["date_updated"]))
            t.StartDate = msToTimePtr(parseInt64Ptr(raw["start_date"]))
            t.DueDate = msToTimePtr(parseInt64Ptr(raw["due_date"]))
            t.TimeSpentMs = parseInt64Ptr(raw["time_spent"]) 

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

            t.TimeEstimate = parseInt64Ptr(raw["time_estimate"])

            if cfArr, ok := raw["custom_fields"].([]interface{}); ok {
                for _, rawCF := range cfArr {
                    cf := rawCF.(map[string]interface{})

                    name, _ := cf["name"].(string)
                    val := cf["value"]
                    if val == nil { continue }

                    // Tanggal Mulai
                    if name == "Tanggal Mulai" {
                        t.StartDate = msToTimePtr(anyToInt64Ptr(val))
                    }

                    // Tanggal Akhir
                    if name == "Tanggal Akhir" {
                        t.DueDate = msToTimePtr(anyToInt64Ptr(val))
                    }
                }
            }

			// Assignees (jamak)
			var assigneeIDs []int64
			if arr, ok := raw["assignees"].([]interface{}); ok {
				for _, assigneeData := range arr {
					if a, ok := assigneeData.(map[string]interface{}); ok {
						if id, ok := a["id"].(float64); ok {
							assigneeIDs = append(assigneeIDs, int64(id))
						}
					}
				}
			}

			if len(assigneeIDs) > 0 {
				firstAssigneeID := assigneeIDs[0]
				t.AssigneeUserID = &firstAssigneeID
			}

            if err := s.Repo.UpsertTask(ctx, t); err != nil {
                log.Println("❌ UPSERT ERROR:", err)
                return total, err
            }
			if err := s.Repo.UpsertTaskAssignees(ctx, t.ID, assigneeIDs); err != nil {
				log.Printf("❌ FAILED TO UPSERT ASSIGNEES for task %s: %v\n", t.ID, err)
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
        if m.Name == t.Username || m.Email == t.Email {
            matchedMember = &m
            break
            }
        }


        fs := model.FullSync{
            TaskID:      t.ID,
            TaskName:    t.Name,
            TaskStatus:  t.Status.Name,
            DateCreated: t.DateCreated,
            DateDone:    t.DateDone,
            DateClosed:  t.DateClosed,
            AssignedTo:  t.Username,
        }

        if matchedMember != nil {
            fs.UserID = matchedMember.ClickUpID
            fs.DisplayName = matchedMember.Name
            fs.Email = matchedMember.Email
            fs.Role = matchedMember.Role
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
        convert := func(t *time.Time) float64 {
            if t == nil {
                return 0
            }
			return float64(t.UnixMilli()) / 1000 / 60 / 60
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
            }

            // Dates
            t.DateCreated = msToTimePtr(toIntPtr(raw["date_created"]))
            t.DateDone = msToTimePtr(toIntPtr(raw["date_done"]))
            t.DateClosed = msToTimePtr(toIntPtr(raw["date_closed"]))
            t.StartDate = msToTimePtr(toIntPtr(raw["start_date"]))
            t.DueDate = msToTimePtr(toIntPtr(raw["due_date"]))
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

func (s *ClickUpService) SyncSpacesAndFolders(ctx context.Context) error {
	if s.TeamID == "" {
		return errors.New("team id not configured")
	}

	// 1. Sync Spaces
	spacesURL := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/space", s.TeamID)
	log.Println("[REQUEST] Syncing Spaces:", spacesURL)
	spaceBytes, err := s.doRequest(ctx, "GET", spacesURL)
	if err != nil {
		return fmt.Errorf("failed to fetch spaces: %w", err)
	}

	var spacesResponse struct {
		Spaces []model.SpaceInfo `json:"spaces"`
	}
	if err := json.Unmarshal(spaceBytes, &spacesResponse); err != nil {
		return fmt.Errorf("failed to parse spaces response: %w", err)
	}

	log.Printf("Found %d spaces. Syncing folders for each space...", len(spacesResponse.Spaces))

	// 2. Sync Folders for each Space
	for _, space := range spacesResponse.Spaces {
		foldersURL := fmt.Sprintf("https://api.clickup.com/api/v2/space/%s/folder", space.ID)
		log.Println("[REQUEST] Syncing Folders for Space", space.ID, ":", foldersURL)
		folderBytes, err := s.doRequest(ctx, "GET", foldersURL)
		if err != nil {
			log.Printf("Could not fetch folders for space %s: %v", space.ID, err)
			continue // Lanjutkan ke space berikutnya jika ada error
		}

		var foldersResponse struct {
			Folders []model.Folder `json:"folders"`
		}
		if err := json.Unmarshal(folderBytes, &foldersResponse); err != nil {
			log.Printf("Could not parse folders for space %s: %v", space.ID, err)
			continue
		}
		log.Printf("Space %s has %d folders.", space.ID, len(foldersResponse.Folders))
		// Di sini Anda bisa menambahkan logika untuk menyimpan data folder ke database jika diperlukan
		// Contoh: s.Repo.UpsertFolders(ctx, foldersResponse.Folders)
	}

	return nil
}

func (s *ClickUpService) AllSync(ctx context.Context) error {
	log.Println("--- STARTING FULL SYNC ---")
    if err := s.SyncTeam(ctx); err != nil {
		return fmt.Errorf("error syncing team: %w", err)
	}
	log.Println("✅ Teams synced successfully.")
	if err := s.SyncSpacesAndFolders(ctx); err != nil {
		return fmt.Errorf("error syncing spaces and folders: %w", err)
	}
	log.Println("✅ Spaces and Folders synced successfully.")
    if err := s.SyncMembers(ctx); err != nil {
		return fmt.Errorf("error syncing members: %w", err)
	}
	log.Println("✅ Members synced successfully.")
	if _, err := s.SyncTasks(ctx); err != nil {
		return fmt.Errorf("error syncing tasks: %w", err)
	}
	log.Println("✅ Tasks synced successfully.")
	log.Println("--- FULL SYNC COMPLETED ---")
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
	return lowerStatus 
}

func (s *ClickUpService) GetWorkload(ctx context.Context, startMs, endMs int64) ([]model.WorkloadUser, error) {
    query := `
        SELECT
            t.id,
            t.name,
            t.status_name,
            t.start_date,
            t.due_date,
            t.date_done,
            t.time_spent_ms,
            u.clickup_id,
            u.name,
            u.username,
            u.email, 
            COALESCE(r.name, '') as role
        FROM tasks t
        JOIN task_assignees ta ON t.id = ta.task_id
        JOIN users u ON ta.user_clickup_id = u.clickup_id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE
            (t.start_date <= to_timestamp($2 / 1000.0) AND t.due_date >= to_timestamp($1 / 1000.0)) OR
            (t.date_done >= to_timestamp($1 / 1000.0) AND t.date_done <= to_timestamp($2 / 1000.0))
        ORDER BY u.name, t.start_date
    `
    rows, err := s.DB.QueryContext(ctx, query, startMs, endMs)
    if err != nil {
        return nil, fmt.Errorf("error querying workload: %w", err)
    }
    defer rows.Close()

    userMap := make(map[int64]*model.WorkloadUser)

    for rows.Next() {
        var taskID, taskName, statusName string
        var startDate, dueDate, dateDone, timeSpentMs sql.NullInt64
        var userID int64
        var userName, userUsername, userEmail, userRole string

        err := rows.Scan(
            &taskID, &taskName, &statusName,
            &startDate, &dueDate, &dateDone, &timeSpentMs,
            &userID, &userName, &userUsername, &userEmail, &userRole,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning workload row: %w", err)
        }

        // Cari atau buat entri pengguna di map
        user, ok := userMap[userID]
        if !ok {
            user = &model.WorkloadUser{
                UserID:        userID,
                Name:          userName,
                Username:      userUsername,
                Email:         userEmail,
                Role:          userRole,
                Tasks:         []model.TaskDetail{},
                ByStatus:      make(map[string]float64),
                ByCategory:    make(map[string]float64),
                StandardHours: 8.0,
            }
            userMap[userID] = user
        }

        // Buat detail tugas
        task := model.TaskDetail{
            ID:         taskID,
            Name:       taskName,
            StatusName: normalizeStatus(statusName),
            StartDate:  msToDateString(toInt64Ptr(startDate)),
            DueDate:    msToDateString(toInt64Ptr(dueDate)),
            DateDone:   msToDateString(toInt64Ptr(dateDone)),
            Assignees:  []model.AssigneeDetail{{ClickUpID: userID, Name: userName, Username: userUsername, Email: userEmail}},
        }
        if timeSpentMs.Valid {
            task.TimeSpentHours = float64(timeSpentMs.Int64) / 3600000.0
        }

        // Tambahkan tugas ke pengguna yang sesuai dan perbarui total jam
        user.Tasks = append(user.Tasks, task)
        user.TotalHours += task.TimeSpentHours
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
			t.description,
			t.text_content,
			t.status_name,
			t.start_date,
			t.due_date,
			t.date_done,
			t.time_spent_ms,
			t.time_estimate
		FROM tasks t
		WHERE
			(t.start_date <= to_timestamp($2 / 1000.0) AND t.due_date >= to_timestamp($1 / 1000.0))
			OR (t.date_done >= to_timestamp($1 / 1000.0) AND t.date_done <= to_timestamp($2 / 1000.0))
	`
	rows, err := s.DB.QueryContext(ctx, query, startMs, endMs) 

	if err != nil {
		log.Printf("Error querying tasks by range: %v", err)
		return nil, err
	}
	defer rows.Close()

	taskMap := make(map[string]*model.TaskDetail)
	var taskOrder []string
	log.Println("--- Scanning Tasks By Range ---")
	for rows.Next() {
		var taskID, taskName, statusName, description, textContent string
		var startDate, dueDate, dateDone sql.NullTime
		var timeSpentMs, timeEstimate sql.NullInt64

		err := rows.Scan(
			&taskID,
			&taskName,
			&description,
			&textContent,
			&statusName,
			&startDate,
			&dueDate,
			&dateDone,
			&timeSpentMs,
			&timeEstimate,
		)
		if err != nil {
			log.Printf("Error scanning task row: %v", err)
			return nil, err
		}

		taskDetail := model.TaskDetail{
			ID:          taskID,
			Name:        taskName,
			Description: description,
			TextContent: textContent,
			StatusName:  statusName,
			StartDate:   nullTimeToDateString(startDate),
			DueDate:     nullTimeToDateString(dueDate),
			DateDone:    nullTimeToDateString(dateDone),
			Assignees:   []model.AssigneeDetail{},
		}

		if timeSpentMs.Valid {
			taskDetail.TimeSpentHours = float64(timeSpentMs.Int64) / 3600000.0
		}

		if timeEstimate.Valid {
			taskDetail.TimeEstimateHours = float64(timeEstimate.Int64) / 3600000.0
			// taskDetail.TimeEstimate sudah tidak ada, jadi baris ini bisa dihapus jika ada
		}

		if _, exists := taskMap[taskID]; !exists {
			taskMap[taskID] = &taskDetail
			taskOrder = append(taskOrder, taskID)
		}
	}
	rows.Close()

	assigneeQuery := `
		SELECT ta.task_id, u.clickup_id, u.username, u.email, u.name
		FROM task_assignees ta
		JOIN users u ON ta.user_clickup_id = u.clickup_id
		WHERE ta.task_id = ANY($1)
	`
	assigneeRows, err := s.DB.QueryContext(ctx, assigneeQuery, pq.Array(taskOrder))
	if err != nil {
		log.Printf("Error querying assignees: %v", err)
		return nil, err
	}
	defer assigneeRows.Close()

	for assigneeRows.Next() {
		var taskID string
		var assignee model.AssigneeDetail
		if err := assigneeRows.Scan(&taskID, &assignee.ClickUpID, &assignee.Username, &assignee.Email, &assignee.Name); err != nil {
			log.Printf("Error scanning assignee row: %v", err)
			return nil, err
		}

		if task, ok := taskMap[taskID]; ok {
			task.Assignees = append(task.Assignees, assignee)
		}
	}
	var result []model.TaskDetail
	for _, taskID := range taskOrder {
		result = append(result, *taskMap[taskID])
	}

	log.Printf("--- Found a total of %d tasks in range ---", len(result))
	return result, nil
}

func nullTimeToDateString(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	s := nt.Time.Format("2006-01-02")
	return &s
}

func toInt64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}