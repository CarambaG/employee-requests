package request

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/CarambaG/employee-requests/internal/domain"
)

const (
	DefaultListLimit = 100
	MaxListLimit     = 1000
	maxDescription   = 10000
)

type CreateParams struct {
	AuthorID    int64
	AssigneeID  int64
	Description string
	DueAt       time.Time
}

type ListFilter struct {
	Status       *domain.RequestStatus
	AssigneeID   *int64
	DepartmentID *int64
	Overdue      *bool
	Limit        int
	Offset       int
}

type Repository interface {
	Create(context.Context, CreateParams) (domain.Request, error)
	GetByNumber(context.Context, int64) (domain.Request, error)
	List(context.Context, ListFilter) ([]domain.Request, error)
	UpdateStatus(context.Context, int64, domain.RequestStatus) (domain.Request, error)
	UpdateAssignee(context.Context, int64, int64) (domain.Request, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (domain.Request, error) {
	params.Description = strings.TrimSpace(params.Description)
	if err := s.validateCreate(params); err != nil {
		return domain.Request{}, err
	}

	return s.repository.Create(ctx, params)
}

func (s *Service) GetByNumber(ctx context.Context, number int64) (domain.Request, error) {
	if number <= 0 {
		return domain.Request{}, fmt.Errorf(
			"%w: request number must be positive",
			domain.ErrInvalidArgument,
		)
	}

	return s.repository.GetByNumber(ctx, number)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]domain.Request, error) {
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if err := validateFilter(filter); err != nil {
		return nil, err
	}

	return s.repository.List(ctx, filter)
}

func (s *Service) UpdateStatus(
	ctx context.Context,
	number int64,
	next domain.RequestStatus,
) (domain.Request, error) {
	if number <= 0 {
		return domain.Request{}, fmt.Errorf(
			"%w: request number must be positive",
			domain.ErrInvalidArgument,
		)
	}
	if !next.IsValid() {
		return domain.Request{}, fmt.Errorf(
			"%w: unknown request status %q",
			domain.ErrInvalidArgument,
			next,
		)
	}

	return s.repository.UpdateStatus(ctx, number, next)
}

func (s *Service) UpdateAssignee(
	ctx context.Context,
	number int64,
	assigneeID int64,
) (domain.Request, error) {
	if number <= 0 {
		return domain.Request{}, fmt.Errorf(
			"%w: request number must be positive",
			domain.ErrInvalidArgument,
		)
	}
	if assigneeID <= 0 {
		return domain.Request{}, fmt.Errorf(
			"%w: assignee_id must be positive",
			domain.ErrInvalidArgument,
		)
	}

	return s.repository.UpdateAssignee(ctx, number, assigneeID)
}

func (s *Service) validateCreate(params CreateParams) error {
	switch {
	case params.AuthorID <= 0:
		return fmt.Errorf("%w: author_id must be positive", domain.ErrInvalidArgument)
	case params.AssigneeID <= 0:
		return fmt.Errorf("%w: assignee_id must be positive", domain.ErrInvalidArgument)
	case params.Description == "":
		return fmt.Errorf("%w: description must not be blank", domain.ErrInvalidArgument)
	case utf8.RuneCountInString(params.Description) > maxDescription:
		return fmt.Errorf(
			"%w: description must not exceed %d characters",
			domain.ErrInvalidArgument,
			maxDescription,
		)
	case params.DueAt.IsZero():
		return fmt.Errorf("%w: due_at is required", domain.ErrInvalidArgument)
	case !params.DueAt.After(s.now()):
		return fmt.Errorf("%w: due_at must be in the future", domain.ErrInvalidArgument)
	default:
		return nil
	}
}

func validateFilter(filter ListFilter) error {
	if filter.Status != nil && !filter.Status.IsValid() {
		return fmt.Errorf("%w: unknown request status", domain.ErrInvalidArgument)
	}
	if filter.AssigneeID != nil && *filter.AssigneeID <= 0 {
		return fmt.Errorf("%w: assignee_id must be positive", domain.ErrInvalidArgument)
	}
	if filter.DepartmentID != nil && *filter.DepartmentID <= 0 {
		return fmt.Errorf("%w: department_id must be positive", domain.ErrInvalidArgument)
	}
	if filter.Limit < 1 || filter.Limit > MaxListLimit {
		return fmt.Errorf(
			"%w: limit must be between 1 and %d",
			domain.ErrInvalidArgument,
			MaxListLimit,
		)
	}
	if filter.Offset < 0 {
		return fmt.Errorf("%w: offset must not be negative", domain.ErrInvalidArgument)
	}

	return nil
}
