package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type Repository interface {
	CreateDepartment(context.Context, string) (domain.Department, error)
	GetDepartmentByID(context.Context, int64) (domain.Department, error)
	ListDepartments(context.Context) ([]domain.Department, error)
	UpdateDepartment(context.Context, int64, string) (domain.Department, error)
	DeleteDepartment(context.Context, int64) error

	CreatePosition(context.Context, string) (domain.Position, error)
	GetPositionByID(context.Context, int64) (domain.Position, error)
	ListPositions(context.Context) ([]domain.Position, error)
	UpdatePosition(context.Context, int64, string) (domain.Position, error)
	DeletePosition(context.Context, int64) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateDepartment(ctx context.Context, name string) (domain.Department, error) {
	name, err := validateName(name)
	if err != nil {
		return domain.Department{}, err
	}

	return s.repository.CreateDepartment(ctx, name)
}

func (s *Service) GetDepartmentByID(ctx context.Context, id int64) (domain.Department, error) {
	if err := validateID(id, "department"); err != nil {
		return domain.Department{}, err
	}

	return s.repository.GetDepartmentByID(ctx, id)
}

func (s *Service) ListDepartments(ctx context.Context) ([]domain.Department, error) {
	return s.repository.ListDepartments(ctx)
}

func (s *Service) UpdateDepartment(
	ctx context.Context,
	id int64,
	name string,
) (domain.Department, error) {
	if err := validateID(id, "department"); err != nil {
		return domain.Department{}, err
	}

	name, err := validateName(name)
	if err != nil {
		return domain.Department{}, err
	}

	return s.repository.UpdateDepartment(ctx, id, name)
}

func (s *Service) DeleteDepartment(ctx context.Context, id int64) error {
	if err := validateID(id, "department"); err != nil {
		return err
	}

	return s.repository.DeleteDepartment(ctx, id)
}

func (s *Service) CreatePosition(ctx context.Context, name string) (domain.Position, error) {
	name, err := validateName(name)
	if err != nil {
		return domain.Position{}, err
	}

	return s.repository.CreatePosition(ctx, name)
}

func (s *Service) GetPositionByID(ctx context.Context, id int64) (domain.Position, error) {
	if err := validateID(id, "position"); err != nil {
		return domain.Position{}, err
	}

	return s.repository.GetPositionByID(ctx, id)
}

func (s *Service) ListPositions(ctx context.Context) ([]domain.Position, error) {
	return s.repository.ListPositions(ctx)
}

func (s *Service) UpdatePosition(
	ctx context.Context,
	id int64,
	name string,
) (domain.Position, error) {
	if err := validateID(id, "position"); err != nil {
		return domain.Position{}, err
	}

	name, err := validateName(name)
	if err != nil {
		return domain.Position{}, err
	}

	return s.repository.UpdatePosition(ctx, id, name)
}

func (s *Service) DeletePosition(ctx context.Context, id int64) error {
	if err := validateID(id, "position"); err != nil {
		return err
	}

	return s.repository.DeletePosition(ctx, id)
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: name must not be blank", domain.ErrInvalidArgument)
	}
	if len([]rune(name)) > 255 {
		return "", fmt.Errorf("%w: name must not exceed 255 characters", domain.ErrInvalidArgument)
	}

	return name, nil
}

func validateID(id int64, resource string) error {
	if id <= 0 {
		return fmt.Errorf("%w: %s id must be positive", domain.ErrInvalidArgument, resource)
	}

	return nil
}
