package domain_test

import (
	"errors"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
)

func TestRequestChangeStatus(t *testing.T) {
	tests := []struct {
		name    string
		current domain.RequestStatus
		next    domain.RequestStatus
		wantErr bool
	}{
		{
			name:    "new to in progress",
			current: domain.RequestStatusNew,
			next:    domain.RequestStatusInProgress,
		},
		{
			name:    "in progress to completed",
			current: domain.RequestStatusInProgress,
			next:    domain.RequestStatusCompleted,
		},
		{
			name:    "new to completed",
			current: domain.RequestStatusNew,
			next:    domain.RequestStatusCompleted,
			wantErr: true,
		},
		{
			name:    "completed to in progress",
			current: domain.RequestStatusCompleted,
			next:    domain.RequestStatusInProgress,
			wantErr: true,
		},
		{
			name:    "unknown status",
			current: domain.RequestStatusNew,
			next:    domain.RequestStatus("unknown"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := domain.Request{Status: tt.current}

			err := request.ChangeStatus(tt.next)
			if tt.wantErr && err == nil {
				t.Fatal("expected an error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && request.Status != tt.next {
				t.Fatalf("unexpected status: got %q, want %q", request.Status, tt.next)
			}
		})
	}
}

func TestRequestChangeStatusReturnsTypedErrors(t *testing.T) {
	request := domain.Request{Status: domain.RequestStatusNew}

	err := request.ChangeStatus(domain.RequestStatusCompleted)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}

	err = request.ChangeStatus(domain.RequestStatus("cancelled"))
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
