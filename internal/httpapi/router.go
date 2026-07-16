package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/employee"
	"github.com/CarambaG/employee-requests/internal/report"
	requestservice "github.com/CarambaG/employee-requests/internal/request"
)

const (
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 10 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 60 * time.Second
)

type Pinger interface {
	Ping(context.Context) error
}

type EmployeeService interface {
	Create(context.Context, employee.CreateParams) (domain.Employee, error)
	GetByID(context.Context, int64) (domain.Employee, error)
	List(context.Context) ([]domain.Employee, error)
	Update(context.Context, int64, employee.UpdateParams) (domain.Employee, error)
	Delete(context.Context, int64) error
}

type RequestService interface {
	Create(context.Context, requestservice.CreateParams) (domain.Request, error)
	GetByNumber(context.Context, int64) (domain.Request, error)
	List(context.Context, requestservice.ListFilter) ([]domain.Request, error)
	UpdateStatus(context.Context, int64, domain.RequestStatus) (domain.Request, error)
	UpdateAssignee(context.Context, int64, int64) (domain.Request, error)
}

type ReportService interface {
	GetSummary(context.Context) (report.Summary, error)
}

type CatalogService interface {
	CreateDepartment(context.Context, string) (domain.Department, error)
	GetDepartmentByID(context.Context, int64) (domain.Department, error)
	ListDepartments(context.Context) ([]domain.Department, error)
	UpdateDepartment(context.Context, int64, string) (domain.Department, error)
	DeleteDepartment(context.Context, int64) error

	CreatePosition(context.Context, string) (domain.Position, error)
	GetPositionByID(context.Context, int64) (domain.Position, error)
	ListPositions(context.Context) ([]domain.Position, error)
	UpdatePosition(context.Context, int64, string) (domain.Position, error)
	DeletePosition(context.Context, int64) error
}

type Dependencies struct {
	Database  Pinger
	Employees EmployeeService
	Catalogs  CatalogService
	Requests  RequestService
	Reports   ReportService
}

func NewRouter(dependencies Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health(dependencies.Database))

	mux.HandleFunc("POST /api/v1/departments", createDepartment(dependencies.Catalogs))
	mux.HandleFunc("GET /api/v1/departments", listDepartments(dependencies.Catalogs))
	mux.HandleFunc("GET /api/v1/departments/{id}", getDepartment(dependencies.Catalogs))
	mux.HandleFunc("PUT /api/v1/departments/{id}", updateDepartment(dependencies.Catalogs))
	mux.HandleFunc("DELETE /api/v1/departments/{id}", deleteDepartment(dependencies.Catalogs))

	mux.HandleFunc("POST /api/v1/positions", createPosition(dependencies.Catalogs))
	mux.HandleFunc("GET /api/v1/positions", listPositions(dependencies.Catalogs))
	mux.HandleFunc("GET /api/v1/positions/{id}", getPosition(dependencies.Catalogs))
	mux.HandleFunc("PUT /api/v1/positions/{id}", updatePosition(dependencies.Catalogs))
	mux.HandleFunc("DELETE /api/v1/positions/{id}", deletePosition(dependencies.Catalogs))

	mux.HandleFunc("POST /api/v1/employees", createEmployee(dependencies.Employees))
	mux.HandleFunc("GET /api/v1/employees", listEmployees(dependencies.Employees))
	mux.HandleFunc("GET /api/v1/employees/{id}", getEmployee(dependencies.Employees))
	mux.HandleFunc("PUT /api/v1/employees/{id}", updateEmployee(dependencies.Employees))
	mux.HandleFunc("DELETE /api/v1/employees/{id}", deleteEmployee(dependencies.Employees))

	mux.HandleFunc("POST /api/v1/requests", createRequest(dependencies.Requests))
	mux.HandleFunc("GET /api/v1/requests", listRequests(dependencies.Requests))
	mux.HandleFunc("GET /api/v1/requests/{number}", getRequest(dependencies.Requests))
	mux.HandleFunc("PATCH /api/v1/requests/{number}/status", updateRequestStatus(dependencies.Requests))
	mux.HandleFunc("PATCH /api/v1/requests/{number}/assignee", updateRequestAssignee(dependencies.Requests))

	mux.HandleFunc("GET /api/v1/reports/summary", getReportSummary(dependencies.Reports))

	return mux
}
