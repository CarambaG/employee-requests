package domain

import "fmt"

type RequestStatus string

const (
	RequestStatusNew        RequestStatus = "new"
	RequestStatusInProgress RequestStatus = "in_progress"
	RequestStatusCompleted  RequestStatus = "completed"
)

func (s RequestStatus) IsValid() bool {
	switch s {
	case RequestStatusNew, RequestStatusInProgress, RequestStatusCompleted:
		return true
	default:
		return false
	}
}

func (s RequestStatus) CanTransitionTo(next RequestStatus) bool {
	switch s {
	case RequestStatusNew:
		return next == RequestStatusInProgress
	case RequestStatusInProgress:
		return next == RequestStatusCompleted
	default:
		return false
	}
}

func (r *Request) ChangeStatus(next RequestStatus) error {
	if !next.IsValid() {
		return fmt.Errorf("unknown request status: %q", next)
	}

	if !r.Status.CanTransitionTo(next) {
		return fmt.Errorf("request status transition %q -> %q is not allowed", r.Status, next)
	}

	r.Status = next
	return nil
}
