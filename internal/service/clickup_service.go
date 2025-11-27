
package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
    "log"
    "strconv"

    "github.com/roksva123/go-kinerja-backend/internal/model"
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
    url := "https://api.clickup.com/api/v2/user"

    b, err := s.doRequest(ctx, "GET", url)
    if err != nil {
        return err
    }

    var out struct {
        User struct {
            ID       int64  `json:"id"`
            Username string `json:"username"`
            Email    string `json:"email"`
            Color    string `json:"color"`
        } `json:"user"`
    }

    if err := json.Unmarshal(b, &out); err != nil {
        return err
    }

    // Mapping ke model
    u := &model.User{
        ID:        out.User.ID,
        ClickUpID: out.User.ID,
        DisplayName:  out.User.Username,
        Name:      out.User.Username,
        Email:     out.User.Email,
        Role:      "employee",
        Color:     out.User.Color,
    }

    if err := s.Repo.UpsertUser(ctx, u); err != nil {
        fmt.Println("ERROR UPSERT USER:", err)
        return err
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
            t := &model.TaskResponse{}

            log.Println("---- PROCESSING TASK ----")

            if id, ok := raw["id"].(string); ok {
                t.ID = id
            }
            log.Println("ID:", t.ID)

            if name, ok := raw["name"].(string); ok {
                t.Name = name
            }
            log.Println("Name:", t.Name)

            if txt, ok := raw["text_content"].(string); ok {
                t.TextContent = txt
            }

            if desc, ok := raw["description"].(string); ok {
                t.Description = desc
            }

            if st, ok := raw["status"].(map[string]interface{}); ok {
                if sid, ok := st["id"].(string); ok {
                    t.Status.ID = sid
                }
                if sname, ok := st["name"].(string); ok {
                    t.Status.Name = sname
                }
                if stype, ok := st["type"].(string); ok {
                    t.Status.Type = stype
                }
                if scol, ok := st["color"].(string); ok {
                    t.Status.Color = scol
                }
            }
            log.Printf("Status: %+v\n", t.Status)

            if dd, ok := raw["date_done"].(string); ok && dd != "" {
                if v, err := strconv.ParseInt(dd, 10, 64); err == nil {
                    t.DateDone = &v
                }
            }
            if t.DateDone != nil {
                log.Println("Date Done:", *t.DateDone)
            } else {
                log.Println("Date Done: NULL")
            }

            if dc, ok := raw["date_closed"].(string); ok && dc != "" {
                if v, err := strconv.ParseInt(dc, 10, 64); err == nil {
                    t.DateClosed = &v
                }
            }
            if t.DateClosed != nil {
                log.Println("Date Closed:", *t.DateClosed)
            } else {
                log.Println("Date Closed: NULL")
            }

            // Start Date
            if sd, ok := raw["start_date"].(string); ok && sd != "" {
                if v, err := strconv.ParseInt(sd, 10, 64); err == nil {
                    t.StartDate = &v
                }
            }

            // Due Date
            if dd, ok := raw["due_date"].(string); ok && dd != "" {
                if v, err := strconv.ParseInt(dd, 10, 64); err == nil {
                    t.DueDate = &v
                }
            }

            // Time Estimate
            if teRaw, ok := raw["time_estimate"]; ok && teRaw != nil {
                switch val := teRaw.(type) {
                case float64:
                    v := int64(val)
                    t.TimeEstimate = &v
                case string:
                    if v, err := strconv.ParseInt(val, 10, 64); err == nil {
                        t.TimeEstimate = &v
                    }
                }
            }

            // Assignees
            if assArr, ok := raw["assignees"].([]interface{}); ok && len(assArr) > 0 {
                if a0, ok := assArr[0].(map[string]interface{}); ok {
                    if uname, ok := a0["username"].(string); ok {
                        t.Username = uname
                    }
                    if email, ok := a0["email"].(string); ok {
                        t.Email = email
                    }
                    if col, ok := a0["color"].(string); ok {
                        t.Color = col
                    }
                }
            }
            log.Printf("Assignee: %s (%s) Color:%s\n", t.Username, t.Email, t.Color)

            // UPSERT to DB
            log.Println("UPSERT TASK:", t.ID)

            if err := s.Repo.UpsertTask(ctx, t); err != nil {
                log.Println("❌ UPSERT ERROR:", err)
                log.Printf("RAW TASK: %+v\n", raw)
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
