package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"

    "github.com/roksva123/go-kinerja-backend/internal/model"
)


type ClickUpUser struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Color    string `json:"color"`
}

type ClickUpTeamsResponse struct {
    Teams []struct {
        ID      string `json:"id"`
        Name    string `json:"name"`
        Members []struct {
            User struct {
                ID       int    `json:"id"`
                Username string `json:"username"`
                Email    string `json:"email"`
                RoleKey  string `json:"role_key"`
            } `json:"user"`
        } `json:"members"`
    } `json:"teams"`
}



// type ClickUpService struct {
//     Token string
//     Repo  UserRepository
// }

// func NewClickUpService(token string, repo UserRepository) *ClickUpService {
//     return &ClickUpService{
//         Token: token,
//         Repo:  repo,
//     }
// }



func (s *ClickUpService) SyncUsersFromClickUp(ctx context.Context) error {
    url := "https://api.clickup.com/api/v2/team"

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", s.Token)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    body, _ := io.ReadAll(res.Body)

    if res.StatusCode != 200 {
        return fmt.Errorf("clickup api error %d: %s", res.StatusCode, string(body))
    }

    var teamRes ClickUpTeamsResponse
    if err := json.Unmarshal(body, &teamRes); err != nil {
        return err
    }

    if len(teamRes.Teams) == 0 {
        return errors.New("no teams found in ClickUp")
    }

    // Loop semua team dan member
    for _, team := range teamRes.Teams {
        for _, member := range team.Members {

            u := member.User

            username := u.Email
            if username == "" {
                username = u.Username
            }

            role := u.RoleKey
            if role == "" {
                role = "employee"
            }

            // SIMPAN KE DATABASE
            err := s.Repo.UpsertUser(ctx, &model.User{
                ID:       int64(u.ID), // FIX
                Username: username,
                Name:     u.Username,
                Role:     role,
            })

            if err != nil {
                return err
            }
        }
    }

    return nil
}
