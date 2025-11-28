package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type WorkloadService struct {
	Repo           *repository.PostgresRepo
	StdHoursPerDay float64
}

func NewWorkloadService(repo *repository.PostgresRepo) *WorkloadService {
	return &WorkloadService{
		Repo:           repo,
		StdHoursPerDay: 8.0,
	}
}

func workingDaysBetween(start, end time.Time) int {
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

func msToHours(ms int64) float64 {
	return float64(ms) / 1000.0 / 3600.0
}

func (s *WorkloadService) BuildWorkload(
	ctx context.Context,
	start, end time.Time,
	positionTag, source, nameFilter string,
) (*model.WorkloadResponse, error) {

	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	users, err := s.Repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	tasks, err := s.Repo.GetTasksByRange(ctx, &startMs, &endMs, nameFilter, "")
	if err != nil {
		return nil, err
	}

	// Group tasks by user
	tasksByUser := map[string][]model.TaskResponse{}
	for _, t := range tasks {
		key := t.Username
		if key == "" {
			key = t.Email
		}
		tasksByUser[key] = append(tasksByUser[key], toTaskResponse(t))
	}

	// Standard hours calc
	wdays := workingDaysBetween(start, end)
	standardHours := float64(wdays) * s.StdHoursPerDay

	resp := &model.WorkloadResponse{
		Start:                 start.Format("2006-01-02"),
		End:                   end.Format("2006-01-02"),
		StandardHoursPerPerson: standardHours,
	}

	totalHoursAll := 0.0
	outUsers := []model.WorkloadUser{}

	for _, u := range users {
		key := u.DisplayName
		if key == "" {
			key = u.Email
		}

		userTasks := tasksByUser[key]
		totalHours := 0.0

		byStatus := map[string]float64{}
		byCategory := map[string]float64{}

		for _, tk := range userTasks {

			ms := int64(0)

			if tk.TimeSpentMs != nil && *tk.TimeSpentMs > 0 {
				ms = *tk.TimeSpentMs
			} else if tk.TimeEstimateMs != nil && *tk.TimeEstimateMs > 0 {
				ms = *tk.TimeEstimateMs
			}

			hours := msToHours(ms)
			if math.IsNaN(hours) || math.IsInf(hours, 0) {
				hours = 0
			}

			totalHours += hours

			status := tk.Status.Name
			if status == "" {
				status = "unknown"
			}

			norm := normalizeStatus(status)
			byStatus[norm] += hours
			byCategory[status] += hours
		}

		totalHoursAll += totalHours

		outUsers = append(outUsers, model.WorkloadUser{
			UserID:        u.ID,
			Username:      u.DisplayName,
			Email:         u.Email,
			Role:          u.Role,
			Color:         u.Color,
			TotalHours:    totalHours,
			TaskCount:     len(userTasks),
			TotalTasks:    int64(len(userTasks)),
			TotalMs:       int64(totalHours * 3600 * 1000),
			Tasks:         userTasks,
			ByStatus:      byStatus,
			ByCategory:    byCategory,
			StandardHours: standardHours,
		})
	}

	resp.Users = outUsers
	resp.Summary.TotalUsers = len(outUsers)
	resp.Summary.TotalHours = totalHoursAll
	if len(outUsers) > 0 {
		resp.Summary.AvgHours = totalHoursAll / float64(len(outUsers))
	}

	return resp, nil
}

func normalizeStatus(s string) string {
	l := strings.ToLower(s)
	switch {
	case strings.Contains(l, "todo") || strings.Contains(l, "to do"):
		return "to-do"
	case strings.Contains(l, "progress") || strings.Contains(l, "in progress") || strings.Contains(l, "inprogress"):
		return "in-progress"
	case strings.Contains(l, "done") || strings.Contains(l, "completed"):
		return "done"
	case strings.Contains(l, "cancel"):
		return "canceled"
	default:
		return l
	}
}

func toTaskResponse(t model.Task) model.TaskResponse {
	return model.TaskResponse{
		ID:          t.TaskID,
		Name:        t.Name,
		TextContent: t.TextContent,
		Description: t.Description,
		Status: struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Type  string `json:"type"`
			Color string `json:"color"`
		}{
			ID:    t.Status.ID,
			Name:  t.Status.Name,
			Type:  t.Status.Type,
			Color: t.Status.Color,
		},
		DateDone:       t.DateDone,
		DateClosed:     t.DateClosed,
		Username:       t.Username,
		Email:          t.Email,
		Color:          t.Color,
		TimeEstimateMs: t.TimeEstimateMs,
		TimeSpentMs:    t.TimeSpentMs,
		StartDate:      t.StartDate,
		DueDate:        t.DueDate,
		TimeEstimate:   t.TimeEstimate,
	}
}
