package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/httpapi"
	"github.com/CarambaG/employee-requests/internal/report"
)

type reportServiceStub struct {
	getSummaryFunc func(context.Context) (report.Summary, error)
}

func (s reportServiceStub) GetSummary(ctx context.Context) (report.Summary, error) {
	return s.getSummaryFunc(ctx)
}

func TestGetReportSummary(t *testing.T) {
	service := reportServiceStub{
		getSummaryFunc: func(context.Context) (report.Summary, error) {
			return report.Summary{
				RequestsByStatus: []report.StatusCount{
					{Status: domain.RequestStatusNew, Count: 10},
					{Status: domain.RequestStatusInProgress, Count: 5},
					{Status: domain.RequestStatusCompleted, Count: 20},
				},
				OverdueRequests: 2,
				CompletedByAssignee: []report.AssigneeCompletedCount{
					{AssigneeID: 7, AssigneeName: "Иванов Иван Иванович", Count: 12},
				},
			}, nil
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/summary", nil)
	recorder := httptest.NewRecorder()

	httpapi.NewRouter(httpapi.Dependencies{
		Database: successfulPinger(),
		Reports:  service,
	}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}

	body := recorder.Body.String()
	for _, fragment := range []string{
		`"requests_by_status"`,
		`"code":"in_progress"`,
		`"name":"В работе"`,
		`"overdue_requests":2`,
		`"completed_requests":12`,
		`"full_name":"Иванов Иван Иванович"`,
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response body does not contain %s: %s", fragment, body)
		}
	}
}
