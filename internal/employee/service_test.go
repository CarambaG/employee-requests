package employee

import (
	"context"
	"errors"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type repositoryStub struct {
	createParams CreateParams
	updateID     int64
	updateParams UpdateParams
	deleteID     int64
}

func (r *repositoryStub) Create(_ context.Context, params CreateParams) (domain.Employee, error) {
	r.createParams = params
	return domain.Employee{ID: 1, FullName: params.FullName}, nil
}

func (r *repositoryStub) GetByID(_ context.Context, id int64) (domain.Employee, error) {
	return domain.Employee{ID: id}, nil
}

func (r *repositoryStub) List(context.Context) ([]domain.Employee, error) {
	return []domain.Employee{{ID: 1}}, nil
}

func (r *repositoryStub) Update(
	_ context.Context,
	id int64,
	params UpdateParams,
) (domain.Employee, error) {
	r.updateID = id
	r.updateParams = params
	return domain.Employee{ID: id, FullName: params.FullName}, nil
}

func (r *repositoryStub) Delete(_ context.Context, id int64) error {
	r.deleteID = id
	return nil
}

func TestCreateTrimsFullName(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	created, err := service.Create(context.Background(), CreateParams{
		FullName:     "  Иванов Иван Иванович  ",
		DepartmentID: 1,
		PositionID:   2,
	})
	if err != nil {
		t.Fatalf("create employee: %v", err)
	}

	if created.FullName != "Иванов Иван Иванович" {
		t.Fatalf("unexpected full name: %q", created.FullName)
	}
	if repository.createParams.FullName != "Иванов Иван Иванович" {
		t.Fatalf("repository received untrimmed name: %q", repository.createParams.FullName)
	}
}

func TestCreateRejectsInvalidEmployee(t *testing.T) {
	tests := []CreateParams{
		{FullName: " ", DepartmentID: 1, PositionID: 1},
		{FullName: "Иванов Иван", DepartmentID: 0, PositionID: 1},
		{FullName: "Иванов Иван", DepartmentID: 1, PositionID: -1},
	}

	service := NewService(&repositoryStub{})
	for _, params := range tests {
		_, err := service.Create(context.Background(), params)
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument, got %v", err)
		}
	}
}

func TestOperationsRejectInvalidID(t *testing.T) {
	service := NewService(&repositoryStub{})

	if _, err := service.GetByID(context.Background(), 0); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("GetByID: expected ErrInvalidArgument, got %v", err)
	}
	if _, err := service.Update(context.Background(), -1, UpdateParams{}); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("Update: expected ErrInvalidArgument, got %v", err)
	}
	if err := service.Delete(context.Background(), 0); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("Delete: expected ErrInvalidArgument, got %v", err)
	}
}
