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

func (s *ClickUpService) SyncUsersFromClickUp(ctx context.Context) error {
	url := "https://api.clickup.com/api/v2/team"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", s.Token)

	res, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return fmt.Errorf("ClickUp API error %d: %s", res.StatusCode, string(body))
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

			// username fallback
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

func (s *ClickUpService) SyncTeam(ctx context.Context) error {

	url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/space", s.TeamID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", s.Token)

	res, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return fmt.Errorf("ClickUp API error %d: %s", res.StatusCode, string(body))
	}

	// PARSE SPACES
	var data struct {
		Spaces []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"spaces"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data.Spaces) == 0 {
		return errors.New("no spaces found for this team")
	}

	// SAVE AS TEAM
	for _, sp := range data.Spaces {

		parent := "" // default: no parent

		err := s.Repo.UpsertTeam(ctx, sp.ID, sp.Name, parent)
		if err != nil {
			return err
		}
	}

	return nil
}



func (s *ClickUpService) FetchTeams() (interface{}, error) {

	url := "https://api.clickup.com/api/v2/team"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", s.Token)

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("ClickUp API error %d: %s", res.StatusCode, string(body))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *ClickUpService) GetTeams(ctx context.Context) ([]model.Team, error) {
	return s.Repo.GetTeams(ctx)
}



func (s *ClickUpService) GetMembers(ctx context.Context) ([]model.User, error) {
    return s.Repo.GetUsers(ctx)
}


func (s *ClickUpService) GetTasks(ctx context.Context) ([]model.Task, error) {
    return s.Repo.GetTasks(ctx)
}
