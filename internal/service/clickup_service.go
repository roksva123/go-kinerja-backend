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
                Email2   string `json:"email"` 
                Username2 string `json:"username"`
                // etc
            } `json:"user"`
        } `json:"members"`
    }
    if err := json.Unmarshal(b, &out); err != nil {
        return err
    }
    for _, m := range out.Members {
        u := &model.User{
            ID:      m.User.ID,
            ClickUpID: m.User.ID,
            Username: m.User.Username,
            Name:    m.User.Username,
            Email:   m.User.Email,
            Role:    "employee",
            Color:   m.User.Color,
        }
        if err := s.Repo.UpsertUser(ctx, u); err != nil {
           fmt.Println("ERROR UPSERT USER:", err)
           return err
        }

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

func (s *ClickUpService) GetTasks(ctx context.Context) ([]model.TaskResponse, error) {
    return s.Repo.GetTasks(ctx)
}

func (s *ClickUpService) GetMembers(ctx context.Context) ([]model.User, error) {
    return s.Repo.GetMembers(ctx)
}

func (s *ClickUpService) GetTeams(ctx context.Context) ([]model.Team, error) {
    return s.Repo.GetTeams(ctx)
}
