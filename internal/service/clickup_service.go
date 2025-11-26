
package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"

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

// SyncTeam -> fetch spaces and save as teams
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

// SyncMembers -> fetch team members and upsert to users
func (s *ClickUpService) SyncMembers(ctx context.Context) error {
    url := "https://api.clickup.com/api/v2/user"

    b, err := s.doRequest(ctx, "GET", url)
    if err != nil {
        return err
    }

    // Struktur response ClickUp untuk endpoint /user
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

    // Mapping ke model kamu
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

// SyncTasks -> paginated tasks from team (note: team task endpoint can be heavy)
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
        var out struct {
            Tasks []map[string]interface{} `json:"tasks"`
        }
        if err := json.Unmarshal(b, &out); err != nil {
            return total, err
        }
        if len(out.Tasks) == 0 {
            break
        }
        for _, raw := range out.Tasks {
            t := &model.TaskResponse{}
            // ID, name
            if id, ok := raw["id"].(string); ok {
                t.ID = id
            }
            if name, ok := raw["name"].(string); ok {
                t.Name = name
            }
            if txt, ok := raw["text_content"].(string); ok {
                t.TextContent = txt
            }
            if desc, ok := raw["description"].(string); ok {
                t.Description = desc
            }
            // status nested
            if st, ok := raw["status"].(map[string]interface{}); ok {
                if sid, ok := st["id"].(string); ok {
                    t.Status.ID = sid
                }
                if sname, ok := st["status"].(string); ok {
                    t.Status.Name = sname
                } else if sname2, ok := st["name"].(string); ok {
                    t.Status.Name = sname2
                }
                if stype, ok := st["type"].(string); ok {
                    t.Status.Type = stype
                }
                if scol, ok := st["color"].(string); ok {
                    t.Status.Color = scol
                }
            }
            // dates (ClickUp returns numeric as string sometimes)
            if dd, ok := raw["date_done"].(float64); ok {
                v := int64(dd)
                t.DateDone = &v
            } else if sdd, ok := raw["date_done"].(string); ok && sdd != "" {
                // try parse numeric string
            }
            if dc, ok := raw["date_closed"].(float64); ok {
                v := int64(dc)
                t.DateClosed = &v
            }
            // assignee or creator or assignees list
            if assArr, ok := raw["assignees"].([]interface{}); ok && len(assArr) > 0 {
                if a0, ok := assArr[0].(map[string]interface{}); ok {
                    if uname, ok := a0["username"].(string); ok {
                        t.Username = uname
                    } else if name, ok := a0["username"].(string); ok {
                        t.Username = name
                    }
                    if email, ok := a0["email"].(string); ok {
                        t.Email = email
                    }
                    if col, ok := a0["color"].(string); ok {
                        t.Color = col
                    }
                }
            } else if creator, ok := raw["creator"].(map[string]interface{}); ok {
                if uname, ok := creator["username"].(string); ok {
                    t.Username = uname
                }
                if email, ok := creator["email"].(string); ok {
                    t.Email = email
                }
                if col, ok := creator["color"].(string); ok {
                    t.Color = col
                }
            }

            // save to DB (upsert)
            if err := s.Repo.UpsertTask(ctx, t); err != nil {
               fmt.Println("ERROR UPSERT TASK:", err)
               fmt.Printf("RAW TASK: %+v\n", raw)
               return total, err
            }

            total++
        }
        page++
    }
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
        // cari user yang match dengan t.Username atau email
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

    // RANGE otomatis (last_6_months, next_6_months)
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