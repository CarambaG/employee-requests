package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/httpapi"
	requestservice "github.com/CarambaG/employee-requests/internal/request"
)

type requestServiceStub struct {
	createFunc         func(context.Context, requestservice.CreateParams) (domain.Request, error)
	getFunc            func(context.Context, int64) (domain.Request, error)
	listFunc           func(context.Context, requestservice.ListFilter) ([]domain.Request, error)
	updateStatusFunc   func(context.Context, int64, domain.RequestStatus) (domain.Request, error)
	updateAssigneeFunc func(context.Context, int64, int64) (domain.Request, error)
}

func (s requestServiceStub) Create(
	ctx context.Context,
	params requestservice.CreateParams,
) (domain.Request, error) {
	return s.createFunc(ctx, params)
}

func (s requestServiceStub) GetByNumber(ctx context.Context, number int64) (domain.Request, error) {
	return s.getFunc(ctx, number)
}

func (s requestServiceStub) List(
	ctx context.Context,
	filter requestservice.ListFilter,
) ([]domain.Request, error) {
	return s.listFunc(ctx, filter)
}

func (s requestServiceStub) UpdateStatus(
	ctx context.Context,
	number int64,
	next domain.RequestStatus,
) (domain.Request, error) {
	return s.updateStatusFunc(ctx, number, next)
}

func (s requestServiceStub) UpdateAssignee(
	ctx context.Context,
	number int64,
	assigneeID int64,
) (domain.Request, error) {
	return s.updateAssigneeFunc(ctx, number, assigneeID)
}

func TestCreateRequest(t *testing.T) {
	dueAt := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	service := requestServiceStub{
		createFunc: func(_ context.Context, params requestservice.CreateParams) (domain.Request, error) {
			if params.AuthorID != 1 || params.AssigneeID != 2 {
				t.Fatalf("unexpected employee ids: author=%d assignee=%d", params.AuthorID, params.AssigneeID)
			}
			if !params.DueAt.Equal(dueAt) {
				t.Fatalf("unexpected due date: %s", params.DueAt)
			}
			return sampleRequest(15), nil
		},
	}

	httpRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/requests",
		strings.NewReader(`{
			"author_id":1,
			"assignee_id":2,
			"description":"Настроить рабочее место",
			"due_at":"2026-07-20T12:00:00Z"
		}`),
	)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(service).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("Location") != "/api/v1/requests/15" {
		t.Fatalf("unexpected Location header: %q", recorder.Header().Get("Location"))
	}
	if !strings.Contains(recorder.Body.String(), `"code":"new"`) ||
		!strings.Contains(recorder.Body.String(), `"name":"Новая"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestListRequestsParsesFilters(t *testing.T) {
	service := requestServiceStub{
		listFunc: func(_ context.Context, filter requestservice.ListFilter) ([]domain.Request, error) {
			if filter.Status == nil || *filter.Status != domain.RequestStatusInProgress {
				t.Fatalf("unexpected status filter: %v", filter.Status)
			}
			if filter.AssigneeID == nil || *filter.AssigneeID != 42 {
				t.Fatalf("unexpected assignee filter: %v", filter.AssigneeID)
			}
			if filter.DepartmentID == nil || *filter.DepartmentID != 7 {
				t.Fatalf("unexpected department filter: %v", filter.DepartmentID)
			}
			if filter.Overdue == nil || !*filter.Overdue {
				t.Fatalf("unexpected overdue filter: %v", filter.Overdue)
			}
			if filter.Limit != 25 || filter.Offset != 50 {
				t.Fatalf("unexpected pagination: limit=%d offset=%d", filter.Limit, filter.Offset)
			}
			return []domain.Request{sampleRequest(15)}, nil
		},
	}

	httpRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/requests?status=in_progress&assignee_id=42&department_id=7&overdue=true&limit=25&offset=50",
		nil,
	)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(service).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.Count(recorder.Body.String(), `"number"`) != 1 {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestListRequestsRejectsInvalidOverdueFilter(t *testing.T) {
	httpRequest := httptest.NewRequest(http.MethodGet, "/api/v1/requests?overdue=yes-please", nil)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(requestServiceStub{}).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"invalid_query"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestGetRequestReturnsNotFound(t *testing.T) {
	service := requestServiceStub{
		getFunc: func(context.Context, int64) (domain.Request, error) {
			return domain.Request{}, domain.ErrNotFound
		},
	}
	httpRequest := httptest.NewRequest(http.MethodGet, "/api/v1/requests/99", nil)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(service).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"request not found"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestUpdateRequestStatus(t *testing.T) {
	service := requestServiceStub{
		updateStatusFunc: func(
			_ context.Context,
			number int64,
			next domain.RequestStatus,
		) (domain.Request, error) {
			if number != 15 || next != domain.RequestStatusInProgress {
				t.Fatalf("unexpected update: number=%d status=%q", number, next)
			}
			updated := sampleRequest(number)
			updated.Status = next
			return updated, nil
		},
	}

	httpRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/requests/15/status",
		strings.NewReader(`{"status":"in_progress"}`),
	)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(service).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"in_progress"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestUpdateRequestStatusRejectsSkippedTransition(t *testing.T) {
	service := requestServiceStub{
		updateStatusFunc: func(
			context.Context,
			int64,
			domain.RequestStatus,
		) (domain.Request, error) {
			return domain.Request{}, errors.Join(domain.ErrConflict, domain.ErrInvalidTransition)
		},
	}

	httpRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/requests/15/status",
		strings.NewReader(`{"status":"completed"}`),
	)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(service).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusConflict)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"invalid_status_transition"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestUpdateRequestAssignee(t *testing.T) {
	service := requestServiceStub{
		updateAssigneeFunc: func(
			_ context.Context,
			number int64,
			assigneeID int64,
		) (domain.Request, error) {
			if number != 15 || assigneeID != 7 {
				t.Fatalf("unexpected update: number=%d assignee=%d", number, assigneeID)
			}
			updated := sampleRequest(number)
			updated.Assignee = sampleEmployee(assigneeID)
			return updated, nil
		},
	}

	httpRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/requests/15/assignee",
		strings.NewReader(`{"assignee_id":7}`),
	)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(service).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"id":7`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRequestEndpointRejectsInvalidNumber(t *testing.T) {
	httpRequest := httptest.NewRequest(http.MethodGet, "/api/v1/requests/not-a-number", nil)
	recorder := httptest.NewRecorder()

	newRouterWithRequests(requestServiceStub{}).ServeHTTP(recorder, httpRequest)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func newRouterWithRequests(requests requestServiceStub) http.Handler {
	return httpapi.NewRouter(httpapi.Dependencies{
		Database:  successfulPinger(),
		Employees: employeeServiceStub{},
		Catalogs:  emptyCatalogService(),
		Requests:  requests,
	})
}

func sampleRequest(number int64) domain.Request {
	createdAt := time.Date(2026, time.July, 16, 9, 0, 0, 0, time.UTC)
	return domain.Request{
		Number:      number,
		CreatedAt:   createdAt,
		Author:      sampleEmployee(1),
		Assignee:    sampleEmployee(2),
		Description: "Настроить рабочее место",
		DueAt:       createdAt.Add(48 * time.Hour),
		Status:      domain.RequestStatusNew,
		UpdatedAt:   createdAt,
	}
}
