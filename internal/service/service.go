package service

import (
	"context"

	"employees-api/internal/domain"
	"employees-api/internal/repository"

	"github.com/google/uuid"
)

type EmployeeService struct {
	repo *repository.EmployeeRepository
}

func NewEmployeeService(repo *repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, req domain.CreateEmployeeRequest) (*domain.Employee, error) {
	validationErrs := &ValidationErrors{}

	req.FullName = NormalizeString(req.FullName)
	if err := ValidateFullName(req.FullName); err != nil {
		validationErrs.Add("fullName", err.Error())
	}

	req.Phone = NormalizeString(req.Phone)
	if err := ValidatePhone(req.Phone); err != nil {
		validationErrs.Add("phone", err.Error())
	}

	req.City = NormalizeString(req.City)
	if err := ValidateCity(req.City); err != nil {
		validationErrs.Add("city", err.Error())
	}

	if validationErrs.HasErrors() {
		return nil, validationErrs
	}

	return s.repo.Create(ctx, req)
}

func (s *EmployeeService) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EmployeeService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}
