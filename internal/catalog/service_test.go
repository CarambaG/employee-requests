package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type repositoryStub struct {
	createDepartmentFunc func(context.Context, string) (domain.Department, error)
	updatePositionFunc   func(context.Context, int64, string) (domain.Position, error)
}

func (s repositoryStub) CreateDepartment(ctx context.Context, name string) (domain.Department, error) {
	return s.createDepartmentFunc(ctx, name)
}
func (repositoryStub) GetDepartmentByID(context.Context, int64) (domain.Department, error) {
	return domain.Department{}, nil
}
func (repositoryStub) ListDepartments(context.Context) ([]domain.Department, error) {
	return nil, nil
}
func (repositoryStub) UpdateDepartment(context.Context, int64, string) (domain.Department, error) {
	return domain.Department{}, nil
}
func (repositoryStub) DeleteDepartment(context.Context, int64) error { return nil }
func (repositoryStub) CreatePosition(context.Context, string) (domain.Position, error) {
	return domain.Position{}, nil
}
func (repositoryStub) GetPositionByID(context.Context, int64) (domain.Position, error) {
	return domain.Position{}, nil
}
func (repositoryStub) ListPositions(context.Context) ([]domain.Position, error) { return nil, nil }
func (s repositoryStub) UpdatePosition(
	ctx context.Context,
	id int64,
	name string,
) (domain.Position, error) {
	return s.updatePositionFunc(ctx, id, name)
}
func (repositoryStub) DeletePosition(context.Context, int64) error { return nil }

func TestCreateDepartmentTrimsName(t *testing.T) {
	service := NewService(repositoryStub{
		createDepartmentFunc: func(_ context.Context, name string) (domain.Department, error) {
			if name != "Разработка" {
				t.Fatalf("unexpected name: %q", name)
			}
			return domain.Department{ID: 1, Name: name}, nil
		},
	})

	created, err := service.CreateDepartment(context.Background(), "  Разработка  ")
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	if created.Name != "Разработка" {
		t.Fatalf("unexpected created department: %+v", created)
	}
}

func TestCreateDepartmentRejectsBlankName(t *testing.T) {
	service := NewService(repositoryStub{})

	_, err := service.CreateDepartment(context.Background(), "   ")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestUpdatePositionRejectsInvalidID(t *testing.T) {
	service := NewService(repositoryStub{})

	_, err := service.UpdatePosition(context.Background(), 0, "Инженер")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
