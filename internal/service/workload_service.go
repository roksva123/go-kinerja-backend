package service

import (
	"context"
	"strings"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
	// "github.com/roksva123/go-kinerja-backend/internal/service"
)

type WorkloadService struct {
	repo       *repository.PostgresRepo
	clickupSvc *ClickUpService
}

func NewWorkloadService(repo *repository.PostgresRepo, clickupSvc *ClickUpService) *WorkloadService {
	return &WorkloadService{
		repo:       repo,
		clickupSvc: clickupSvc,
	}
}

func (s *WorkloadService) GetTasksSummary(ctx context.Context, startDate, endDate time.Time, name, email string) ([]model.TaskSummary, error) {
	summaries, err := s.repo.GetTasksSummaryByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if name == "" && email == "" {
		return summaries, nil
	}

	var filteredSummaries []model.TaskSummary
	for _, summary := range summaries {
		nameMatch := name == "" || strings.Contains(strings.ToLower(summary.Name), strings.ToLower(name))
		emailMatch := email == "" || strings.Contains(strings.ToLower(summary.Email), strings.ToLower(email))

		if nameMatch && emailMatch {
			filteredSummaries = append(filteredSummaries, summary)
		}
	}

	return filteredSummaries, nil
}

func (s *WorkloadService) GetWorkload(ctx context.Context, start, end time.Time, username string) ([]model.WorkloadUser, error) {
	workloads, err := s.repo.GetWorkload(ctx, start, end)
	if err != nil {
		return nil, err
	}

	if username == "" {
		return workloads, nil
	}

	var filteredWorkloads []model.WorkloadUser
	for _, workload := range workloads {
		if strings.Contains(strings.ToLower(workload.Username), strings.ToLower(username)) {
			filteredWorkloads = append(filteredWorkloads, workload)
		}
	}

	return filteredWorkloads, nil
}

func (s *WorkloadService) GetTasksByRangeGroupedByAssignee(ctx context.Context, start, end time.Time, sortOrder string) (map[int64]model.AssigneeDetail, map[int64][]model.TaskDetail, error) {
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	tasks, err := s.repo.GetTasksFull(ctx, &startMs, &endMs, "", "", "")
	if err != nil {
		return nil, nil, err
	}

	assigneesMap := make(map[int64]model.AssigneeDetail)
	tasksByAssignee := make(map[int64][]model.TaskDetail)

	for _, task := range tasks {
		var userID int64

		if task.UserID == nil {
			userID = 0 
			if _, ok := assigneesMap[userID]; !ok {
				assigneesMap[userID] = model.AssigneeDetail{
					ClickUpID: userID,
					Username:  "Unassigned",
					Name:      "Unassigned",
				}
			}
		} else {
			userID = *task.UserID
		}

		if _, ok := assigneesMap[userID]; !ok {
			var username, email, name string
			if task.Username != nil {
				username = *task.Username
				name = *task.Username 
			}
			if task.Email != nil {
				email = *task.Email
			}

			assigneesMap[userID] = model.AssigneeDetail{
				ClickUpID: userID,
				Username:  username,
				Email:     email,
				Name:      name,
			}
		}

		tasksByAssignee[userID] = append(tasksByAssignee[userID], model.TaskDetail{
			ID:   task.TaskID,
			Name: task.TaskName,
			Description: task.Description,
			StatusName: task.StatusName,
			ProjectName: task.ProjectName,
			StartDate: nullTimeToDateStringPointer(task.StartDate),
			DueDate: nullTimeToDateStringPointer(task.DueDate),
			DateDone: nullTimeToDateStringPointer(task.DateDone),
			TimeEstimateHours: task.TimeEstimateHours,
			TimeSpentHours: task.TimeSpentHours,
		})
	}

	return assigneesMap, tasksByAssignee, nil
}

func nullTimeToDateStringPointer(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("02-01-2006")
	return &s
}