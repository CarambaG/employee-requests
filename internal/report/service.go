package report

import (
	"context"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type StatusCount struct {
	Status domain.RequestStatus `json:"status"`
	Count  int64                `json:"count"`
}

type AssigneeCompletedCount struct {
	AssigneeID   int64  `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
	Count        int64  `json:"count"`
}

type Summary struct {
	RequestsByStatus    []StatusCount
	OverdueRequests     int64
	CompletedByAssignee []AssigneeCompletedCount
}

type Repository interface {
	GetSummary(context.Context) (Summary, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetSummary(ctx context.Context) (Summary, error) {
	return s.repository.GetSummary(ctx)
}
