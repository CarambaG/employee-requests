package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/employee"
	"github.com/CarambaG/employee-requests/internal/httpapi"
)

type pingerFunc func(context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

type employeeServiceStub struct {
	createFunc func(context.Context, employee.CreateParams) (domain.Employee, error)
	getFunc    func(context.Context, int64) (domain.Employee, error)
	listFunc   func(context.Context) ([]domain.Employee, error)
	updateFunc func(context.Context, int64, employee.UpdateParams) (domain.Employee, error)
	deleteFunc func(context.Context, int64) error
}

func (s employeeServiceStub) Create(
	ctx context.Context,
	params employee.CreateParams,
) (domain.Employee, error) {
	return s.createFunc(ctx, params)
}

func (s employeeServiceStub) GetByID(ctx context.Context, id int64) (domain.Employee, error) {
	return s.getFunc(ctx, id)
}

func (s employeeServiceStub) List(ctx context.Context) ([]domain.Employee, error) {
	return s.listFunc(ctx)
}

func (s employeeServiceStub) Update(
	ctx context.Context,
	id int64,
	params employee.UpdateParams,
) (domain.Employee, error) {
	return s.updateFunc(ctx, id, params)
}

func (s employeeServiceStub) Delete(ctx context.Context, id int64) error {
	return s.deleteFunc(ctx, id)
}

func TestHealthReturnsOKWhenDatabaseIsAvailable(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	newTestRouter(employeeServiceStub{}, pingerFunc(func(context.Context) error {
		return nil
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"database":"available"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestHealthReturnsUnavailableWhenDatabasePingFails(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	newTestRouter(employeeServiceStub{}, pingerFunc(func(context.Context) error {
		return errors.New("database is unavailable")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"unexpected status code: got %d, want %d",
			recorder.Code,
			http.StatusServiceUnavailable,
		)
	}
}

func TestCreateEmployee(t *testing.T) {
	service := employeeServiceStub{
		createFunc: func(_ context.Context, params employee.CreateParams) (domain.Employee, error) {
			if params.FullName != "Иванов Иван Иванович" {
				t.Fatalf("unexpected full name: %q", params.FullName)
			}
			return sampleEmployee(7), nil
		},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/employees",
		strings.NewReader(`{"full_name":"Иванов Иван Иванович","department_id":1,"position_id":2}`),
	)
	recorder := httptest.NewRecorder()

	newTestRouter(service, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("Location") != "/api/v1/employees/7" {
		t.Fatalf("unexpected Location header: %q", recorder.Header().Get("Location"))
	}
	if !strings.Contains(recorder.Body.String(), `"full_name":"Иванов Иван Иванович"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestCreateEmployeeRejectsInvalidJSON(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/employees",
		strings.NewReader(`{"full_name":`),
	)
	recorder := httptest.NewRecorder()

	newTestRouter(employeeServiceStub{}, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"invalid_json"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestGetEmployeeReturnsNotFound(t *testing.T) {
	service := employeeServiceStub{
		getFunc: func(context.Context, int64) (domain.Employee, error) {
			return domain.Employee{}, domain.ErrNotFound
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/employees/42", nil)
	recorder := httptest.NewRecorder()

	newTestRouter(service, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"not_found"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestListEmployees(t *testing.T) {
	service := employeeServiceStub{
		listFunc: func(context.Context) ([]domain.Employee, error) {
			return []domain.Employee{sampleEmployee(1), sampleEmployee(2)}, nil
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
	recorder := httptest.NewRecorder()

	newTestRouter(service, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.Count(recorder.Body.String(), `"full_name"`) != 2 {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestUpdateEmployeeReturnsConflict(t *testing.T) {
	service := employeeServiceStub{
		updateFunc: func(context.Context, int64, employee.UpdateParams) (domain.Employee, error) {
			return domain.Employee{}, domain.ErrConflict
		},
	}
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/employees/7",
		strings.NewReader(`{"full_name":"Иванов Иван Иванович","department_id":999,"position_id":2}`),
	)
	recorder := httptest.NewRecorder()

	newTestRouter(service, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func TestDeleteEmployee(t *testing.T) {
	var deletedID int64
	service := employeeServiceStub{
		deleteFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/employees/7", nil)
	recorder := httptest.NewRecorder()

	newTestRouter(service, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if deletedID != 7 {
		t.Fatalf("unexpected deleted employee id: %d", deletedID)
	}
}

func TestEmployeeEndpointRejectsInvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/employees/not-a-number", nil)
	recorder := httptest.NewRecorder()

	newTestRouter(employeeServiceStub{}, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func newTestRouter(service employeeServiceStub, database httpapi.Pinger) http.Handler {
	return httpapi.NewRouter(httpapi.Dependencies{
		Database:  database,
		Employees: service,
	})
}

func successfulPinger() pingerFunc {
	return func(context.Context) error { return nil }
}

func sampleEmployee(id int64) domain.Employee {
	return domain.Employee{
		ID:       id,
		FullName: "Иванов Иван Иванович",
		Department: domain.Department{
			ID:   1,
			Name: "Разработка",
		},
		Position: domain.Position{
			ID:   2,
			Name: "Инженер-программист",
		},
	}
}
