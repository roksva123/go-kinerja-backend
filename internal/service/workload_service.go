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
	// Placeholder implementation
	return nil, nil, nil
}