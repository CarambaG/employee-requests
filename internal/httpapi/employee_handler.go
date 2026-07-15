package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/employee"
)

type employeeWriteRequest struct {
	FullName     string `json:"full_name"`
	DepartmentID int64  `json:"department_id"`
	PositionID   int64  `json:"position_id"`
}

type employeeResponse struct {
	ID         int64             `json:"id"`
	FullName   string            `json:"full_name"`
	Department referenceResponse `json:"department"`
	Position   referenceResponse `json:"position"`
}

type referenceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func createEmployee(service EmployeeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request employeeWriteRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		created, err := service.Create(r.Context(), employee.CreateParams{
			FullName:     request.FullName,
			DepartmentID: request.DepartmentID,
			PositionID:   request.PositionID,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v1/employees/%d", created.ID))
		writeJSON(w, http.StatusCreated, newEmployeeResponse(created))
	}
}

func getEmployee(service EmployeeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := employeeID(w, r)
		if !ok {
			return
		}

		found, err := service.GetByID(r.Context(), id)
		if err != nil {
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newEmployeeResponse(found))
	}
}

func listEmployees(service EmployeeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employees, err := service.List(r.Context())
		if err != nil {
			writeServiceError(w, err)
			return
		}

		response := make([]employeeResponse, 0, len(employees))
		for _, found := range employees {
			response = append(response, newEmployeeResponse(found))
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func updateEmployee(service EmployeeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := employeeID(w, r)
		if !ok {
			return
		}

		var request employeeWriteRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		updated, err := service.Update(r.Context(), id, employee.UpdateParams{
			FullName:     request.FullName,
			DepartmentID: request.DepartmentID,
			PositionID:   request.PositionID,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newEmployeeResponse(updated))
	}
}

func deleteEmployee(service EmployeeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := employeeID(w, r)
		if !ok {
			return
		}

		if err := service.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func employeeID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_id", "employee id must be a positive integer")
		return 0, false
	}

	return id, true
}

func newEmployeeResponse(found domain.Employee) employeeResponse {
	return employeeResponse{
		ID:       found.ID,
		FullName: found.FullName,
		Department: referenceResponse{
			ID:   found.Department.ID,
			Name: found.Department.Name,
		},
		Position: referenceResponse{
			ID:   found.Position.ID,
			Name: found.Position.Name,
		},
	}
}
