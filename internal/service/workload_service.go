package service

import (
	"context"
	"time"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"github.com/roksva123/go-kinerja-backend/internal/repository"
)

type WorkloadService struct {
	Repo           *repository.PostgresRepo
	ClickSvc       *ClickUpService 
	StdHoursPerDay float64
}


func NewWorkloadService(repo *repository.PostgresRepo, clickSvc *ClickUpService) *WorkloadService {
	return &WorkloadService{
		Repo:           repo,
		ClickSvc:       clickSvc,
		StdHoursPerDay: 8.0,
	}
}

func msToHours(ms int64) float64 {
	return float64(ms) / 1000.0 / 3600.0
}

func (s *WorkloadService) GetTasksByRangeGroupedByAssignee(ctx context.Context, startDate, endDate time.Time, sortOrder string) (map[int64]model.AssigneeDetail, map[int64][]model.TaskDetail, error) {
	tasks, err := s.ClickSvc.GetTasksByRange(ctx, startDate.UnixMilli(), endDate.UnixMilli(), sortOrder)
	if err != nil {
		return nil, nil, err
	}

	tasksByAssigneeMap := make(map[int64][]model.TaskDetail)
	assigneesMap := make(map[int64]model.AssigneeDetail)

	for _, task := range tasks {
		if task.Assignees == nil || len(task.Assignees) == 0 {
			continue
		}

		for _, assignee := range task.Assignees {
			if _, ok := assigneesMap[assignee.ClickUpID]; !ok {
				assigneesMap[assignee.ClickUpID] = assignee
			}

			tasksByAssigneeMap[assignee.ClickUpID] = append(tasksByAssigneeMap[assignee.ClickUpID], task)
		}
	}

	return assigneesMap, tasksByAssigneeMap, nil
}
