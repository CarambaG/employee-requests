package employee

import (
	"context"
	"fmt"
	"strings"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type CreateParams struct {
	FullName     string
	DepartmentID int64
	PositionID   int64
}

type UpdateParams struct {
	FullName     string
	DepartmentID int64
	PositionID   int64
}

type Repository interface {
	Create(context.Context, CreateParams) (domain.Employee, error)
	GetByID(context.Context, int64) (domain.Employee, error)
	List(context.Context) ([]domain.Employee, error)
	Update(context.Context, int64, UpdateParams) (domain.Employee, error)
	Delete(context.Context, int64) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (domain.Employee, error) {
	params.FullName = strings.TrimSpace(params.FullName)
	if err := validateEmployee(params.FullName, params.DepartmentID, params.PositionID); err != nil {
		return domain.Employee{}, err
	}

	return s.repository.Create(ctx, params)
}

func (s *Service) GetByID(ctx context.Context, id int64) (domain.Employee, error) {
	if err := validateID(id); err != nil {
		return domain.Employee{}, err
	}

	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Employee, error) {
	return s.repository.List(ctx)
}

func (s *Service) Update(
	ctx context.Context,
	id int64,
	params UpdateParams,
) (domain.Employee, error) {
	if err := validateID(id); err != nil {
		return domain.Employee{}, err
	}

	params.FullName = strings.TrimSpace(params.FullName)
	if err := validateEmployee(params.FullName, params.DepartmentID, params.PositionID); err != nil {
		return domain.Employee{}, err
	}

	return s.repository.Update(ctx, id, params)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func validateEmployee(fullName string, departmentID, positionID int64) error {
	if fullName == "" {
		return fmt.Errorf("%w: full_name must not be blank", domain.ErrInvalidArgument)
	}
	if departmentID <= 0 {
		return fmt.Errorf("%w: department_id must be positive", domain.ErrInvalidArgument)
	}
	if positionID <= 0 {
		return fmt.Errorf("%w: position_id must be positive", domain.ErrInvalidArgument)
	}

	return nil
}

func validateID(id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: employee id must be positive", domain.ErrInvalidArgument)
	}

	return nil
}
