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
type UserRepository interface {
    UpsertUser(ctx context.Context, user *model.User) error
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
func (s *ClickUpService) SyncUsersFromClickUp(ctx context.Context) error {
    url := "https://api.clickup.com/api/v2/team"

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", s.Token)
    fmt.Println(">>> Sending Header:", req.Header)        

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

            err := s.Repo.UpsertUser(ctx, &model.User{
                ID:       int64(u.ID),
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

func (c *ClickUpService) FetchTeams() (interface{}, error) {
    url := "https://api.clickup.com/api/v2/team"

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", c.Token)

    res, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    body, _ := io.ReadAll(res.Body)
    if res.StatusCode >= 300 {
        return nil, fmt.Errorf("clickup api error %d: %s", res.StatusCode, string(body))
    }

    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, err
    }

    return data, nil
}
