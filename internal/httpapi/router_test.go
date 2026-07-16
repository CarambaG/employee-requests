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

type catalogServiceStub struct {
	createDepartmentFunc func(context.Context, string) (domain.Department, error)
	getDepartmentFunc    func(context.Context, int64) (domain.Department, error)
	listDepartmentsFunc  func(context.Context) ([]domain.Department, error)
	updateDepartmentFunc func(context.Context, int64, string) (domain.Department, error)
	deleteDepartmentFunc func(context.Context, int64) error
	createPositionFunc   func(context.Context, string) (domain.Position, error)
	getPositionFunc      func(context.Context, int64) (domain.Position, error)
	listPositionsFunc    func(context.Context) ([]domain.Position, error)
	updatePositionFunc   func(context.Context, int64, string) (domain.Position, error)
	deletePositionFunc   func(context.Context, int64) error
}

func (s catalogServiceStub) CreateDepartment(ctx context.Context, name string) (domain.Department, error) {
	return s.createDepartmentFunc(ctx, name)
}
func (s catalogServiceStub) GetDepartmentByID(ctx context.Context, id int64) (domain.Department, error) {
	return s.getDepartmentFunc(ctx, id)
}
func (s catalogServiceStub) ListDepartments(ctx context.Context) ([]domain.Department, error) {
	return s.listDepartmentsFunc(ctx)
}
func (s catalogServiceStub) UpdateDepartment(ctx context.Context, id int64, name string) (domain.Department, error) {
	return s.updateDepartmentFunc(ctx, id, name)
}
func (s catalogServiceStub) DeleteDepartment(ctx context.Context, id int64) error {
	return s.deleteDepartmentFunc(ctx, id)
}
func (s catalogServiceStub) CreatePosition(ctx context.Context, name string) (domain.Position, error) {
	return s.createPositionFunc(ctx, name)
}
func (s catalogServiceStub) GetPositionByID(ctx context.Context, id int64) (domain.Position, error) {
	return s.getPositionFunc(ctx, id)
}
func (s catalogServiceStub) ListPositions(ctx context.Context) ([]domain.Position, error) {
	return s.listPositionsFunc(ctx)
}
func (s catalogServiceStub) UpdatePosition(ctx context.Context, id int64, name string) (domain.Position, error) {
	return s.updatePositionFunc(ctx, id, name)
}
func (s catalogServiceStub) DeletePosition(ctx context.Context, id int64) error {
	return s.deletePositionFunc(ctx, id)
}

func emptyCatalogService() catalogServiceStub {
	return catalogServiceStub{
		createDepartmentFunc: func(context.Context, string) (domain.Department, error) { return domain.Department{}, nil },
		getDepartmentFunc:    func(context.Context, int64) (domain.Department, error) { return domain.Department{}, nil },
		listDepartmentsFunc:  func(context.Context) ([]domain.Department, error) { return nil, nil },
		updateDepartmentFunc: func(context.Context, int64, string) (domain.Department, error) { return domain.Department{}, nil },
		deleteDepartmentFunc: func(context.Context, int64) error { return nil },
		createPositionFunc:   func(context.Context, string) (domain.Position, error) { return domain.Position{}, nil },
		getPositionFunc:      func(context.Context, int64) (domain.Position, error) { return domain.Position{}, nil },
		listPositionsFunc:    func(context.Context) ([]domain.Position, error) { return nil, nil },
		updatePositionFunc:   func(context.Context, int64, string) (domain.Position, error) { return domain.Position{}, nil },
		deletePositionFunc:   func(context.Context, int64) error { return nil },
	}
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
	return newTestRouterWithCatalog(service, emptyCatalogService(), database)
}

func newTestRouterWithCatalog(
	employees employeeServiceStub,
	catalogs catalogServiceStub,
	database httpapi.Pinger,
) http.Handler {
	return httpapi.NewRouter(httpapi.Dependencies{
		Database:  database,
		Employees: employees,
		Catalogs:  catalogs,
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

func TestCreateDepartment(t *testing.T) {
	catalogs := emptyCatalogService()
	catalogs.createDepartmentFunc = func(_ context.Context, name string) (domain.Department, error) {
		if name != "Разработка" {
			t.Fatalf("unexpected department name: %q", name)
		}
		return domain.Department{ID: 3, Name: name}, nil
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/departments",
		strings.NewReader(`{"name":"Разработка"}`),
	)
	recorder := httptest.NewRecorder()

	newTestRouterWithCatalog(employeeServiceStub{}, catalogs, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("Location") != "/api/v1/departments/3" {
		t.Fatalf("unexpected Location header: %q", recorder.Header().Get("Location"))
	}
	if !strings.Contains(recorder.Body.String(), `"name":"Разработка"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestListPositions(t *testing.T) {
	catalogs := emptyCatalogService()
	catalogs.listPositionsFunc = func(context.Context) ([]domain.Position, error) {
		return []domain.Position{
			{ID: 1, Name: "Инженер"},
			{ID: 2, Name: "Руководитель"},
		}, nil
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/positions", nil)
	recorder := httptest.NewRecorder()

	newTestRouterWithCatalog(employeeServiceStub{}, catalogs, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.Count(recorder.Body.String(), `"name"`) != 2 {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestDeleteDepartmentReturnsConflict(t *testing.T) {
	catalogs := emptyCatalogService()
	catalogs.deleteDepartmentFunc = func(context.Context, int64) error {
		return domain.ErrConflict
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/departments/1", nil)
	recorder := httptest.NewRecorder()

	newTestRouterWithCatalog(employeeServiceStub{}, catalogs, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func TestGetPositionReturnsResourceSpecificNotFound(t *testing.T) {
	catalogs := emptyCatalogService()
	catalogs.getPositionFunc = func(context.Context, int64) (domain.Position, error) {
		return domain.Position{}, domain.ErrNotFound
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/positions/42", nil)
	recorder := httptest.NewRecorder()

	newTestRouterWithCatalog(employeeServiceStub{}, catalogs, successfulPinger()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"position not found"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
