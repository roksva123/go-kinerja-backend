package service

import (

	"time"

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
