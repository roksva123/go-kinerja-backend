package service

import (
	"github.com/roksva123/go-kinerja-backend/internal/model"

	"gorm.io/gorm"
)

type EmployeeService struct {
	DB *gorm.DB
}

func NewEmployeeService(db *gorm.DB) *EmployeeService {
	return &EmployeeService{DB: db}
}

func (s *EmployeeService) SyncClickUpUsers(users []ClickUpUser) error {
	for _, u := range users {
		emp := model.Employee{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Color:    u.Color,
		}

		s.DB.Where("id = ?", emp.ID).FirstOrCreate(&emp)
	}
	return nil
}

func (s *EmployeeService) GetAllEmployees() ([]model.Employee, error) {
	var emps []model.Employee
	err := s.DB.Find(&emps).Error
	return emps, err
}
