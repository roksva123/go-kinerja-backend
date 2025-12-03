package service

import (
	"strconv"

	"github.com/roksva123/go-kinerja-backend/internal/model"
	"gorm.io/gorm"
)

type EmployeeService struct {
	DB *gorm.DB
}

func NewEmployeeService(db *gorm.DB) *EmployeeService {
	return &EmployeeService{DB: db}
}

func (s *EmployeeService) SyncEmployee(user model.User) error {

	emp := model.Employee{
		ID:       strconv.FormatInt(user.ClickUpID, 10),
		Name: user.Name,
		Email:    user.Email,
		Color:    "", 
	}

	return s.DB.
		Where("id = ?", emp.ID).
		FirstOrCreate(&emp).
		Error
}

func (s *EmployeeService) GetAllEmployees() ([]model.Employee, error) {
	var emps []model.Employee
	err := s.DB.Find(&emps).Error
	return emps, err
}
