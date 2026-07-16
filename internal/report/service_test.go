package report

import (
	"context"
	"errors"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type repositoryStub struct {
	getSummaryFunc func(context.Context) (Summary, error)
}

func (s repositoryStub) GetSummary(ctx context.Context) (Summary, error) {
	return s.getSummaryFunc(ctx)
}

func TestServiceGetSummary(t *testing.T) {
	expected := Summary{
		RequestsByStatus: []StatusCount{
			{Status: domain.RequestStatusNew, Count: 12},
		},
		OverdueRequests: 3,
	}

	service := NewService(repositoryStub{
		getSummaryFunc: func(context.Context) (Summary, error) {
			return expected, nil
		},
	})

	actual, err := service.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if actual.OverdueRequests != expected.OverdueRequests {
		t.Fatalf(
			"unexpected overdue count: got %d, want %d",
			actual.OverdueRequests,
			expected.OverdueRequests,
		)
	}
}

func TestServiceGetSummaryPropagatesRepositoryError(t *testing.T) {
	expectedError := errors.New("database error")
	service := NewService(repositoryStub{
		getSummaryFunc: func(context.Context) (Summary, error) {
			return Summary{}, expectedError
		},
	})

	_, err := service.GetSummary(context.Background())
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
