package service

import (
	"context"
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

func (s *WorkloadService) GetTasksSummary(ctx context.Context, startMs, endMs *int64, username string) (*model.TaskSummary, error) {
	return s.repo.GetTasksSummary(ctx, startMs, endMs, username)
}

// func (s *WorkloadService) SyncAll(ctx context.Context) error {
// 	// Placeholder implementation
// 	return nil
// }

func (s *WorkloadService) GetTasksByRangeGroupedByAssignee(ctx context.Context, start, end time.Time, sortOrder string) (map[int64]model.AssigneeDetail, map[int64][]model.TaskDetail, error) {
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()

	// Panggil repository untuk mendapatkan semua tugas dalam rentang waktu
	tasks, err := s.repo.GetTasksFull(ctx, &startMs, &endMs, "", "", "")
	if err != nil {
		return nil, nil, err
	}

	assigneesMap := make(map[int64]model.AssigneeDetail)
	tasksByAssignee := make(map[int64][]model.TaskDetail)

	for _, task := range tasks {
		if task.UserID == nil {
			continue // Lewati tugas yang tidak memiliki assignee
		}

		userID := *task.UserID

		// Jika assignee belum ada di map, tambahkan
		if _, ok := assigneesMap[userID]; !ok {
			var username, email, name string
			if task.Username != nil {
				username = *task.Username
				name = *task.Username // Default name to username
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