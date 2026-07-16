package request

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type repositoryStub struct {
	createFunc         func(context.Context, CreateParams) (domain.Request, error)
	getFunc            func(context.Context, int64) (domain.Request, error)
	listFunc           func(context.Context, ListFilter) ([]domain.Request, error)
	updateStatusFunc   func(context.Context, int64, domain.RequestStatus) (domain.Request, error)
	updateAssigneeFunc func(context.Context, int64, int64) (domain.Request, error)
}

func (r repositoryStub) Create(ctx context.Context, params CreateParams) (domain.Request, error) {
	return r.createFunc(ctx, params)
}

func (r repositoryStub) GetByNumber(ctx context.Context, number int64) (domain.Request, error) {
	return r.getFunc(ctx, number)
}

func (r repositoryStub) List(ctx context.Context, filter ListFilter) ([]domain.Request, error) {
	return r.listFunc(ctx, filter)
}

func (r repositoryStub) UpdateStatus(
	ctx context.Context,
	number int64,
	next domain.RequestStatus,
) (domain.Request, error) {
	return r.updateStatusFunc(ctx, number, next)
}

func (r repositoryStub) UpdateAssignee(
	ctx context.Context,
	number int64,
	assigneeID int64,
) (domain.Request, error) {
	return r.updateAssigneeFunc(ctx, number, assigneeID)
}

func TestCreateTrimsDescription(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	repository := repositoryStub{
		createFunc: func(_ context.Context, params CreateParams) (domain.Request, error) {
			if params.Description != "Настроить рабочее место" {
				t.Fatalf("unexpected description: %q", params.Description)
			}
			return domain.Request{Number: 1}, nil
		},
	}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	created, err := service.Create(context.Background(), CreateParams{
		AuthorID:    1,
		AssigneeID:  2,
		Description: "  Настроить рабочее место  ",
		DueAt:       now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if created.Number != 1 {
		t.Fatalf("unexpected request number: %d", created.Number)
	}
}

func TestCreateRejectsPastDueDate(t *testing.T) {
	now := time.Date(2026, time.July, 16, 10, 0, 0, 0, time.UTC)
	service := NewService(repositoryStub{})
	service.now = func() time.Time { return now }

	_, err := service.Create(context.Background(), CreateParams{
		AuthorID:    1,
		AssigneeID:  2,
		Description: "Описание",
		DueAt:       now.Add(-time.Minute),
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestListAppliesDefaultLimit(t *testing.T) {
	repository := repositoryStub{
		listFunc: func(_ context.Context, filter ListFilter) ([]domain.Request, error) {
			if filter.Limit != DefaultListLimit {
				t.Fatalf("unexpected limit: %d", filter.Limit)
			}
			return nil, nil
		},
	}

	if _, err := NewService(repository).List(context.Background(), ListFilter{}); err != nil {
		t.Fatalf("list requests: %v", err)
	}
}

func TestListRejectsUnknownStatus(t *testing.T) {
	status := domain.RequestStatus("cancelled")
	_, err := NewService(repositoryStub{}).List(context.Background(), ListFilter{
		Status: &status,
		Limit:  DefaultListLimit,
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUpdateStatusDelegatesToRepository(t *testing.T) {
	repository := repositoryStub{
		updateStatusFunc: func(
			_ context.Context,
			number int64,
			next domain.RequestStatus,
		) (domain.Request, error) {
			if number != 15 {
				t.Fatalf("unexpected request number: %d", number)
			}
			if next != domain.RequestStatusInProgress {
				t.Fatalf("unexpected target status: %q", next)
			}
			return domain.Request{Number: number, Status: next}, nil
		},
	}

	updated, err := NewService(repository).UpdateStatus(
		context.Background(),
		15,
		domain.RequestStatusInProgress,
	)
	if err != nil {
		t.Fatalf("update request status: %v", err)
	}
	if updated.Status != domain.RequestStatusInProgress {
		t.Fatalf("unexpected status: %q", updated.Status)
	}
}

func TestUpdateStatusRejectsUnknownStatus(t *testing.T) {
	_, err := NewService(repositoryStub{}).UpdateStatus(
		context.Background(),
		15,
		domain.RequestStatus("cancelled"),
	)
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUpdateAssigneeDelegatesToRepository(t *testing.T) {
	repository := repositoryStub{
		updateAssigneeFunc: func(
			_ context.Context,
			number int64,
			assigneeID int64,
		) (domain.Request, error) {
			if number != 15 || assigneeID != 7 {
				t.Fatalf("unexpected update parameters: number=%d assignee=%d", number, assigneeID)
			}
			return domain.Request{
				Number:   number,
				Assignee: domain.Employee{ID: assigneeID},
			}, nil
		},
	}

	updated, err := NewService(repository).UpdateAssignee(context.Background(), 15, 7)
	if err != nil {
		t.Fatalf("update request assignee: %v", err)
	}
	if updated.Assignee.ID != 7 {
		t.Fatalf("unexpected assignee: %d", updated.Assignee.ID)
	}
}

func TestUpdateAssigneeRejectsInvalidID(t *testing.T) {
	_, err := NewService(repositoryStub{}).UpdateAssignee(context.Background(), 15, 0)
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
