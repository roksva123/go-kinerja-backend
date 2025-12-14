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
}

func NewClickUpService(
    repo *repository.PostgresRepo,
    apiKey string,
    token string,
    teamID string,
) *ClickUpService {

    return &ClickUpService{
        Repo:   repo,
        APIKey: apiKey,
        Token:  token,
        TeamID: teamID,
        Client: &http.Client{Timeout: 20 * time.Second},
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

		roleMap := map[string]string{
			"Adi Nugroho":            "infra",
			"Aditya Permadi":         "backend",
			"Alfitra Fadjri":         "web",
			"Andhika Adjie Pradhana": "backend",
			"Andri":                  "web",
			"Arif Hidayat":           "mobile apps",
			"aufa":                   "backend",
			"Christian Wibisono":     "pm",
			"Deni Candra":            "pm",
			"Dwi A Sobarna":          "mobile apps",
			"Egin Tia Yulanda":       "web",
			"Fahri kurniawan":        "backend-web",
			"Heru Septiadi":          "analis",
			"Nurmian Petronella":     "analis",
			"Sani Rosa":              "UI-UX",
		}

		for _, member := range team.Members {
			u := &model.User{
				ClickUpID:   member.User.ID,
				Name:        member.User.Username,
				Email:       member.User.Email,
				Status:      "aktif",
				Role:        roleMap[member.User.Username],
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

func getTimePtr(v interface{}) *time.Time {
    if v == nil {
        return nil
    }

    // Try parsing as milliseconds first
    if ms := parseInt64Ptr(v); ms != nil {
        return msToTimePtr(ms)
    }

    if s, ok := v.(string); ok && s != "" {
        // Try "DD-MM-YYYY" format (from your example JSON)
        if t, err := time.Parse("02-01-2006", s); err == nil {
            return &t
        }
        // Try "YYYY-MM-DD" format (common in databases/other APIs)
        if t, err := time.Parse("2006-01-02", s); err == nil {
            return &t
        }
    }
    return nil
}

// New helper to parse time values (milliseconds or hours) to *float64 (in hours)
func parseTimeValueToHoursPtr(v interface{}) *float64 {
    if v == nil {
        return nil
    }

    // Try parsing as milliseconds (int64 or float64 representing int64)
    if ms := parseInt64Ptr(v); ms != nil {
        hours := float64(*ms) / 3600000.0 // Convert milliseconds to hours
        return &hours
    }

    // If not milliseconds, try parsing as float64 directly (assuming it's already in hours)
    if val, ok := v.(float64); ok {
        return &val
    }
    if s, ok := v.(string); ok && s != "" {
        if x, err := strconv.ParseFloat(s, 64); err == nil {
            return &x
        }
    }
    return nil
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
		url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/task?page=%d&subtasks=true&include_closed=true", s.TeamID, page)
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
			t := &model.TaskResponse{}

			// STEP 1: PARSE ALL PRIMARY DATA
			t.ID = safeString(raw["id"])
			t.Name = safeString(raw["name"])
			t.TextContent = safeString(raw["text_content"])
			t.Description = safeString(raw["description"])

			if st, ok := raw["status"].(map[string]interface{}); ok {
				t.Status.ID = safeString(st["id"])
				t.Status.Name = safeString(st["status"])
				t.Status.Type = safeString(st["type"])
				t.Status.Color = safeString(st["color"])
			}

			if list, ok := raw["list"].(map[string]interface{}); ok {
				if id, ok := list["id"].(string); ok {
					t.ListID = &id
				}
			}

			// Parse all dates from primary source
			t.DateCreated = getTimePtr(raw["date_created"])
			t.DateUpdated = getTimePtr(raw["date_updated"])
			t.DateDone = getTimePtr(raw["date_done"])
			t.DateClosed = getTimePtr(raw["date_closed"])
			t.StartDate = getTimePtr(raw["start_date"])
			t.DueDate = getTimePtr(raw["due_date"])

			// STEP 2: PARSE CUSTOM FIELDS (OVERWRITES DATES IF PRESENT)
			if cfArr, ok := raw["custom_fields"].([]interface{}); ok {
				for _, rawCF := range cfArr {
					cf := rawCF.(map[string]interface{})
					name, _ := cf["name"].(string)
					val := cf["value"]
					if val == nil {
						continue
					}
					if name == "Tanggal Mulai" {
						if newStartDate := getTimePtr(val); newStartDate != nil {
							t.StartDate = newStartDate
						}
					}
					if name == "Tanggal Akhir" {
						if newDueDate := getTimePtr(val); newDueDate != nil {
							t.DueDate = newDueDate
						}
					}
				}
			}

			// STEP 3: CALCULATE EFFICIENCY METRICS (MUST BE DONE BEFORE FALLBACKS)
			if t.StartDate != nil && t.DueDate != nil && t.DateDone != nil {
				log.Printf("Calculating efficiency for Task ID %s with Start: %v, Due: %v, Done: %v", t.ID, *t.StartDate, *t.DueDate, *t.DateDone)
				remainingDuration := t.DueDate.Sub(*t.DateDone)
				remainingHours := remainingDuration.Hours()
				t.RemainingTimeHours = &remainingHours

				durasiAlokasi := t.DueDate.Sub(*t.StartDate)
				durasiAktual := t.DateDone.Sub(*t.StartDate)

				if durasiAktual.Hours() > 0 {
					efficiency := (durasiAlokasi.Hours() / durasiAktual.Hours()) * 100
					t.TimeEfficiencyPercentage = &efficiency
				}
			}

			// STEP 4: APPLY FALLBACK LOGIC FOR MISSING VALUES
			// Fallback for StartDate
			if t.StartDate == nil {
				t.StartDate = t.DateCreated
			}
			// Fallback for DueDate
			if t.DueDate == nil {
				if t.DateDone != nil {
					t.DueDate = t.DateDone
				} else {
					t.DueDate = t.StartDate
				}
			}

			// Fallback for TimeSpent
			if timeSpentHours := parseTimeValueToHoursPtr(raw["time_spent"]); timeSpentHours != nil {
				t.TimeSpentHours = timeSpentHours
			} else if t.TimeSpentHours == nil { // Only calculate if not set
				isDone := t.Status.Type == "done" || t.Status.Type == "closed"
				if isDone && t.StartDate != nil && t.DateDone != nil && t.DateDone.After(*t.StartDate) {
					workingDays := WorkingDaysBetween(*t.StartDate, *t.DateDone)
					hours := float64(workingDays * 8)
					t.TimeSpentHours = &hours
				}
			}
			// Fallback for TimeEstimate
			if timeEstimateHours := parseTimeValueToHoursPtr(raw["time_estimate"]); timeEstimateHours != nil {
				t.TimeEstimateHours = timeEstimateHours
			} else if t.TimeEstimateHours == nil { // Only set default if not set
				defaultHours := 8.0
				t.TimeEstimateHours = &defaultHours
			}

			// STEP 5: PROCESS AND SAVE TO DATABASE
			// Upsert Status
			if t.Status.ID != "" {
				status := model.TaskStatus{
					ID:    t.Status.ID,
					Name:  t.Status.Name,
					Type:  t.Status.Type,
					Color: t.Status.Color,
				}
				if err := s.Repo.UpsertTaskStatus(ctx, &status); err != nil {
					log.Printf("WARNING: Failed to upsert task status %s: %v\n", t.Status.ID, err)
				}
			}

			// Process Assignees
			var assigneeIDs []int64
			if assigneesArr, ok := raw["assignees"].([]interface{}); ok {
				for _, assigneeData := range assigneesArr {
					if a, ok := assigneeData.(map[string]interface{}); ok {
						if id, ok := a["id"].(float64); ok {
							assigneeIDs = append(assigneeIDs, int64(id))
						}
					}
				}
			}
			// Upsert Task to DB
			if err := s.Repo.UpsertTask(ctx, t); err != nil {
				log.Println("❌ UPSERT ERROR:", err)
				return total, err
			}
			// Upsert Assignees relation
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

func (s *ClickUpService) GetSpaces(ctx context.Context) ([]model.SpaceInfo, error) {
	return s.Repo.GetSpaces(ctx)
}

func (s *ClickUpService) GetLists(ctx context.Context) ([]model.List, error) {
	return s.Repo.GetLists(ctx)
}

func (s *ClickUpService) GetFolders(ctx context.Context) ([]model.Folder, error) {
	return s.Repo.GetFolders(ctx)
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
                    if cid, ok := a["id"].(float64); ok {
                        v := int64(cid)
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

func (s *ClickUpService) SyncFolders(ctx context.Context, spaceID string) error {
	foldersURL := fmt.Sprintf("https://api.clickup.com/api/v2/space/%s/folder", spaceID)
	folderBytes, err := s.doRequest(ctx, "GET", foldersURL)
	if err != nil {
		return fmt.Errorf("could not fetch folders for space %s: %w", spaceID, err)
	}

	var foldersResponse struct {
		Folders []model.Folder `json:"folders"`
	}
	if err := json.Unmarshal(folderBytes, &foldersResponse); err != nil {
		return fmt.Errorf("could not parse folders for space %s: %w", spaceID, err)
	}

	for _, folder := range foldersResponse.Folders {
		if err := s.Repo.UpsertFolder(ctx, &folder); err != nil {
			log.Printf("Failed to upsert folder %s: %v", folder.ID, err)
		}
	}
	return nil
}

func nullTimeToDateString(nt sql.NullTime) *string {
	if !nt.Valid {
		return nil
	}
	s := nt.Time.Format("02-01-2006")
	return &s
}

func toInt64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func normalizeStatus(status string) string {
	lowerStatus := strings.ToLower(status)
	if strings.Contains(lowerStatus, "review") || strings.Contains(lowerStatus, "progress") {
		return "progres"
	}
	if strings.Contains(lowerStatus, "do") {
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

func (s *ClickUpService) SyncSpacesAndFolders(ctx context.Context) error {
	if s.TeamID == "" {
		return errors.New("team id not configured")
	}

	// 1. Fetch and Upsert Spaces
	spaces, err := s.fetchSpaces(ctx)
	if err != nil {
		return err
	}

	for _, space := range spaces {
		log.Printf("--- Processing Space: %s (%s) ---", space.Name, space.ID)
		if err := s.Repo.UpsertSpace(ctx, &space); err != nil {
			log.Printf("ERROR: Failed to upsert space %s: %v", space.ID, err)
			continue // Continue to the next space if there's an error
		}

		// 2. Fetch and Upsert Folders within the Space
		folders, err := s.fetchFoldersForSpace(ctx, space.ID)
		if err != nil {
			log.Printf("WARNING: Could not fetch folders for space %s: %v", space.ID, err)
		} else {
			for _, folder := range folders {
				folder.Space.ID = space.ID // Ensure relation to space is correct
				if err := s.Repo.UpsertFolder(ctx, &folder); err != nil {
					log.Printf("ERROR: Failed to upsert folder %s: %v", folder.ID, err)
					continue
				}

				// 3. Fetch and Upsert Lists within each Folder
				lists, err := s.fetchListsForFolder(ctx, folder.ID)
				if err != nil {
					log.Printf("WARNING: Could not fetch lists for folder %s: %v", folder.ID, err)
				} else {
					for i := range lists {
						lists[i].FolderID = folder.ID
						lists[i].SpaceID = space.ID
					}
					if err := s.upsertLists(ctx, lists); err != nil {
						log.Printf("ERROR: Failed to upsert lists for folder %s: %v", folder.ID, err)
					}
				}
			}
		}

		// 4. Fetch and Upsert Folderless Lists within the Space
		folderlessLists, err := s.fetchFolderlessListsForSpace(ctx, space.ID)
		if err != nil {
			log.Printf("WARNING: Could not fetch folderless lists for space %s: %v", space.ID, err)
		} else {
			for i := range folderlessLists {
				folderlessLists[i].SpaceID = space.ID
			}
			if err := s.upsertLists(ctx, folderlessLists); err != nil {
				log.Printf("ERROR: Failed to upsert folderless lists for space %s: %v", space.ID, err)
			}
		}
	}

	return nil
}

// Helper functions to make SyncSpacesAndFolders cleaner

func (s *ClickUpService) fetchSpaces(ctx context.Context) ([]model.SpaceInfo, error) {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/space?archived=false", s.TeamID)
	bytes, err := s.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spaces: %w", err)
	}
	var resp struct {
		Spaces []model.SpaceInfo `json:"spaces"`
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse spaces: %w", err)
	}
	log.Printf("Found %d active spaces.", len(resp.Spaces))
	return resp.Spaces, nil
}

func (s *ClickUpService) GetWorkload(ctx context.Context, startMs, endMs int64) ([]model.WorkloadUser, error) {
	return s.Repo.GetWorkload(ctx, time.UnixMilli(startMs), time.UnixMilli(endMs))
}

func msToDateString(ms *int64) *string { 
	if ms == nil {
		return nil
	}
	t := time.UnixMilli(*ms)
	s := t.Format("2006-01-02")
	return &s
}

func (s *ClickUpService) fetchFoldersForSpace(ctx context.Context, spaceID string) ([]model.Folder, error) {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/space/%s/folder?archived=false", spaceID)
	bytes, err := s.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Folders []model.Folder `json:"folders"`
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, err
	}
	log.Printf("Space %s has %d active folders.", spaceID, len(resp.Folders))
	return resp.Folders, nil
}

func (s *ClickUpService) fetchListsForFolder(ctx context.Context, folderID string) ([]model.List, error) {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/folder/%s/list?archived=false", folderID)
	bytes, err := s.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Lists []model.List `json:"lists"`
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, err
	}
	log.Printf("Folder %s has %d active lists.", folderID, len(resp.Lists))
	return resp.Lists, nil
}

func (s *ClickUpService) fetchFolderlessListsForSpace(ctx context.Context, spaceID string) ([]model.List, error) {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/space/%s/list?archived=false", spaceID)
	bytes, err := s.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Lists []model.List `json:"lists"`
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, err
	}
	log.Printf("Space %s has %d active folderless lists.", spaceID, len(resp.Lists))
	return resp.Lists, nil
}

func (s *ClickUpService) upsertLists(ctx context.Context, lists []model.List) error {
	for _, list := range lists {
		if err := s.Repo.UpsertList(ctx, &list); err != nil {
			log.Printf("ERROR: Failed to upsert list %s: %v", list.ID, err)
		}
	}
	return nil
}

func (s *ClickUpService) GetTasksByRange(ctx context.Context, startMs, endMs int64, sortOrder string) ([]model.TaskDetail, error) {
	orderDirection := "DESC" // Default: terbaru
	if strings.ToLower(sortOrder) == "asc" {
		orderDirection = "ASC" // Terlama
	}

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
			t.time_spent_hours,
			t.time_estimate_hours,
			COALESCE(l.name, f.name) as project_name
		FROM tasks t
		LEFT JOIN lists l ON t.list_id = l.id
		LEFT JOIN folders f ON l.folder_id = f.id
		WHERE
			(t.start_date <= to_timestamp($2 / 1000.0) AND t.due_date >= to_timestamp($1 / 1000.0))
			OR (t.date_done >= to_timestamp($1 / 1000.0) AND t.date_done <= to_timestamp($2 / 1000.0))
		ORDER BY t.start_date %s
	`
	rows, err := s.Repo.DB.QueryContext(ctx, fmt.Sprintf(query, orderDirection), startMs, endMs)

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

		var timeSpentHours, timeEstimateHours sql.NullFloat64
		var projectName sql.NullString

		err := rows.Scan(
			&taskID,
			&taskName,
			&description,
			&textContent,
			&statusName,
			&startDate,
			&dueDate,
			&dateDone,
			&timeSpentHours,
			&timeEstimateHours,
			&projectName,
		)
		if err != nil {
			log.Printf("Error scanning task row: %v", err)
			return nil, err
		}

		taskDetail := model.TaskDetail{ // Variabel taskDetail dibuat di sini
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

		// --- START: Kalkulasi Efisiensi Waktu & Sisa Waktu ---
		if startDate.Valid && dueDate.Valid && dateDone.Valid {
			// 1. Hitung sisa waktu (due_date - date_done)
			remainingDuration := dueDate.Time.Sub(dateDone.Time)
			remainingHoursValue := remainingDuration.Hours()
			taskDetail.RemainingTimeHours = &remainingHoursValue

			// 2. Hitung persentase efisiensi
			durasiAlokasi := dueDate.Time.Sub(startDate.Time)
			durasiAktual := dateDone.Time.Sub(startDate.Time)

			if durasiAktual.Hours() > 0 {
				efficiency := (durasiAlokasi.Hours() / durasiAktual.Hours()) * 100
				taskDetail.TimeEfficiencyPercentage = &efficiency
			}
		}
		// --- END: Kalkulasi ---

		if projectName.Valid {
			taskDetail.ProjectName = &projectName.String
		}

		if timeSpentHours.Valid { taskDetail.TimeSpentHours = timeSpentHours.Float64 }
		if timeEstimateHours.Valid { taskDetail.TimeEstimateHours = timeEstimateHours.Float64 }


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
	assigneeRows, err := s.Repo.DB.QueryContext(ctx, assigneeQuery, pq.Array(taskOrder))
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


func (s *ClickUpService) AllSync(ctx context.Context) error {
	// Call the progress-based function with a nil channel, as we don't need to track progress here.
	return s.AllSyncWithProgress(ctx, nil)
}


func (s *ClickUpService) AllSyncWithProgress(ctx context.Context, progressChan chan<- string) error {
	// Helper function to safely send progress messages
	sendProgress := func(msg string) {
		log.Println(msg) // Also log to console
		if progressChan != nil {
			progressChan <- msg
		}
	}

	sendProgress("--- Starting Full Sync ---")

	sendProgress("Syncing spaces, folders, and lists...")
	if err := s.SyncSpacesAndFolders(ctx); err != nil {
		return fmt.Errorf("error syncing spaces and folders: %w", err)
	}
	sendProgress("✔ Spaces, folders, and lists synced.")

	sendProgress("Syncing members...")
	if err := s.SyncMembers(ctx); err != nil {
		return fmt.Errorf("error syncing members: %w", err)
	}
	sendProgress("✔ Members synced.")

	sendProgress("Syncing tasks...")
	if _, err := s.SyncTasks(ctx); err != nil {
		return fmt.Errorf("error syncing tasks: %w", err)
	}
	sendProgress("✔ Tasks synced.")

	sendProgress("--- Full Sync Completed Successfully ---")
	return nil
}



func WorkingDaysBetween(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}

	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)

	days := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			days++
		}
	}
	return days
}

func (s *ClickUpService) GetTasksByAssignee(ctx context.Context, startMs, endMs int64) (*model.TasksByAssigneeResponse, error) {
	start := time.UnixMilli(startMs)
	end := time.UnixMilli(endMs)

	summaries, err := s.Repo.GetTasksSummaryByDateRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get task summaries: %w", err)
	}

	if len(summaries) == 0 {
		return &model.TasksByAssigneeResponse{
			Count:     0,
			Assignees: []model.AssigneeWithTasks{},
		}, nil
	}

	var assignees []model.AssigneeWithTasks
	for _, summary := range summaries {
		tasks, err := s.Repo.GetTasksByUser(ctx, summary.UserID, start, end)
		if err != nil {
			log.Printf("WARNING: could not get tasks for user %d: %v", summary.UserID, err)
			continue 
		}

		totalSpentHours := summary.TotalSpentHours

		assignee := model.AssigneeWithTasks{
			ClickUpID:          summary.UserID,
			Username:           summary.Name,
			Email:              summary.Email,
			Name:               summary.Name,
			TotalSpentHours:    totalSpentHours,
			ExpectedHours:      summary.TotalWorkHours,
			TotalTasks:         summary.TotalTasks,
			TotalUpcomingHours: summary.TotalUpcomingHours,
			Tasks:              tasks,
		}
		assignees = append(assignees, assignee)
	}

	response := &model.TasksByAssigneeResponse{
		Count:     len(assignees),
		Assignees: assignees,
	}

	return response, nil
}