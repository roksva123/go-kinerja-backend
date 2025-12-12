package service

import (
	"context"
	"strings"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"

	// "github.com/roksva123/go-kinerja-backend/internal/service"
	"fmt"
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

func (s *WorkloadService) GetTasksByRangeGrouped(ctx context.Context, start, end time.Time, sortOrder string) (*model.TasksByAssigneeResponse, error) {
	return s.GetTasksByAssignee(ctx, start, end)
}

func nullTimeToDateStringPointer(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("02-01-2006")
	return &s
}

func (s *WorkloadService) GetTasksByAssignee(ctx context.Context, start, end time.Time) (*model.TasksByAssigneeResponse, error) {
	summaries, err := s.repo.GetTasksSummaryByDateRange(ctx, start, end)
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
		tasks, err := s.repo.GetTasksByUser(ctx, summary.UserID, start, end)
		if err != nil {
			fmt.Printf("WARNING: could not get tasks for user %d: %v\n", summary.UserID, err)
		}

		totalSpentHours := summary.ActualWorkHours

		assignee := model.AssigneeWithTasks{
			ClickUpID:          summary.UserID,
			Username:           summary.Name,
			Email:              summary.Email,
			Name:               summary.Name,
			TotalSpentHours:    totalSpentHours,
			ExpectedHours:      summary.TotalWorkHours,
			TotalTasks:         summary.TotalTasks,
			ActualWorkHours:    summary.ActualWorkHours,
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