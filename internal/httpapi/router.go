package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/employee"
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

type Dependencies struct {
	Database  Pinger
	Employees EmployeeService
}

func NewRouter(dependencies Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health(dependencies.Database))
	mux.HandleFunc("POST /api/v1/employees", createEmployee(dependencies.Employees))
	mux.HandleFunc("GET /api/v1/employees", listEmployees(dependencies.Employees))
	mux.HandleFunc("GET /api/v1/employees/{id}", getEmployee(dependencies.Employees))
	mux.HandleFunc("PUT /api/v1/employees/{id}", updateEmployee(dependencies.Employees))
	mux.HandleFunc("DELETE /api/v1/employees/{id}", deleteEmployee(dependencies.Employees))

	return mux
}
