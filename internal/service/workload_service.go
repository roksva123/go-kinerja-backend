package service

import (
	"context"
	"math"
	"time"
	"strings"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type WorkloadService struct {
	Repo *repository.PostgresRepo
	StdHoursPerDay float64
}

func NewWorkloadService(repo *repository.PostgresRepo) *WorkloadService {
	return &WorkloadService{Repo: repo, StdHoursPerDay: 8.0}
}

// Count weekdays between start and end inclusive
func workingDaysBetween(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)
	days := 0
	for d := start; !d.After(end); d = d.AddDate(0,0,1) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			days++
		}
	}
	return days
}

// ms -> hours float
func msToHours(ms int64) float64 {
	return float64(ms) / 1000.0 / 3600.0
}

func (s *WorkloadService) BuildWorkload(ctx context.Context, start, end time.Time, positionTag, source, nameFilter string) (*model.WorkloadResponse, error) {
	// prepare timestamps in ms
	var fromMs, toMs *int64
	f := start.UnixNano() / int64(time.Millisecond)
	t := end.UnixNano() / int64(time.Millisecond)
	fromMs = &f
	toMs = &t

	// fetch all users (show all users)
	users, err := s.Repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// fetch tasks in range with filters
	tasks, err := s.Repo.GetTasksByRange(ctx, fromMs, toMs, positionTag, source, nameFilter)
	if err != nil {
		return nil, err
	}

	// map username -> tasks
	tasksByUser := map[string][]model.TaskResponse{}
	for _, tk := range tasks {
		key := tk.Username
		// if username empty use email as fallback
		if key == "" {
			key = tk.Email
		}
		tasksByUser[key] = append(tasksByUser[key], tk)
	}

	// standard hours per person
	wdays := workingDaysBetween(start, end)
	standardHours := float64(wdays) * s.StdHoursPerDay

	resp := &model.WorkloadResponse{
		Start: start.Format("2006-01-02"),
		End:   end.Format("2006-01-02"),
	}
	resp.StandardHoursPerPerson = standardHours

	totalHoursAll := 0.0
	usersOut := make([]model.WorkloadUser, 0, len(users))
	for _, u := range users {
		key := u.DisplayName
		if key == "" {
			key = u.Email
		}
		userTasks := tasksByUser[key]
		totalHours := 0.0
		byStatus := map[string]float64{}
		byCat := map[string]float64{}

		for _, tk := range userTasks {
			// how many ms
			ms := int64(0)
			if tk.TimeSpentMs != nil && *tk.TimeSpentMs > 0 {
				ms = *tk.TimeSpentMs
			} else if tk.TimeEstimateMs != nil && *tk.TimeEstimateMs > 0 {
				ms = *tk.TimeEstimateMs
			} else {
				ms = 0
			}
			hours := msToHours(ms)
			if math.IsNaN(hours) || math.IsInf(hours,0) {
				hours = 0
			}
			totalHours += hours

			statusName := tk.Status.Name
			if statusName == "" {
				statusName = "unknown"
			}
			norm := normalizeStatus(statusName)
			byStatus[norm] += hours

			byCat[statusName] += hours
		}

		totalHoursAll += totalHours

		uOut := model.WorkloadUser{
			UserID:       u.ID,
			Username:     u.DisplayName,
			Name:         u.Name,
			Email:        u.Email,
			Role:         u.Role,
			Color:        u.Color,
			TotalHours:   totalHours,
			TaskCount:    len(userTasks),
			Tasks:        userTasks,
			ByStatus:     byStatus,
			ByCategory:   byCat,
			StandardHours: standardHours,
		}
		usersOut = append(usersOut, uOut)
	}

	resp.Users = usersOut
	resp.Summary.TotalUsers = len(usersOut)
	resp.Summary.TotalHours = totalHoursAll
	if resp.Summary.TotalUsers > 0 {
		resp.Summary.AvgHours = totalHoursAll / float64(resp.Summary.TotalUsers)
	}

	return resp, nil
}

// normalize status names to your buckets
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
