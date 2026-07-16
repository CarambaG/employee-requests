package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/CarambaG/employee-requests/internal/domain"
)

type catalogWriteRequest struct {
	Name string `json:"name"`
}

type catalogResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func createDepartment(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request catalogWriteRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		created, err := service.CreateDepartment(r.Context(), request.Name)
		if err != nil {
			writeServiceError(w, err, "department")
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v1/departments/%d", created.ID))
		writeJSON(w, http.StatusCreated, newDepartmentResponse(created))
	}
}

func getDepartment(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := catalogID(w, r, "department")
		if !ok {
			return
		}

		found, err := service.GetDepartmentByID(r.Context(), id)
		if err != nil {
			writeServiceError(w, err, "department")
			return
		}

		writeJSON(w, http.StatusOK, newDepartmentResponse(found))
	}
}

func listDepartments(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		departments, err := service.ListDepartments(r.Context())
		if err != nil {
			writeServiceError(w, err, "department")
			return
		}

		response := make([]catalogResponse, 0, len(departments))
		for _, department := range departments {
			response = append(response, newDepartmentResponse(department))
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func updateDepartment(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := catalogID(w, r, "department")
		if !ok {
			return
		}

		var request catalogWriteRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		updated, err := service.UpdateDepartment(r.Context(), id, request.Name)
		if err != nil {
			writeServiceError(w, err, "department")
			return
		}

		writeJSON(w, http.StatusOK, newDepartmentResponse(updated))
	}
}

func deleteDepartment(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := catalogID(w, r, "department")
		if !ok {
			return
		}

		if err := service.DeleteDepartment(r.Context(), id); err != nil {
			writeServiceError(w, err, "department")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func createPosition(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request catalogWriteRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		created, err := service.CreatePosition(r.Context(), request.Name)
		if err != nil {
			writeServiceError(w, err, "position")
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v1/positions/%d", created.ID))
		writeJSON(w, http.StatusCreated, newPositionResponse(created))
	}
}

func getPosition(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := catalogID(w, r, "position")
		if !ok {
			return
		}

		found, err := service.GetPositionByID(r.Context(), id)
		if err != nil {
			writeServiceError(w, err, "position")
			return
		}

		writeJSON(w, http.StatusOK, newPositionResponse(found))
	}
}

func listPositions(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions, err := service.ListPositions(r.Context())
		if err != nil {
			writeServiceError(w, err, "position")
			return
		}

		response := make([]catalogResponse, 0, len(positions))
		for _, position := range positions {
			response = append(response, newPositionResponse(position))
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func updatePosition(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := catalogID(w, r, "position")
		if !ok {
			return
		}

		var request catalogWriteRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		updated, err := service.UpdatePosition(r.Context(), id, request.Name)
		if err != nil {
			writeServiceError(w, err, "position")
			return
		}

		writeJSON(w, http.StatusOK, newPositionResponse(updated))
	}
}

func deletePosition(service CatalogService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := catalogID(w, r, "position")
		if !ok {
			return
		}

		if err := service.DeletePosition(r.Context(), id); err != nil {
			writeServiceError(w, err, "position")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func catalogID(w http.ResponseWriter, r *http.Request, resource string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			resource+" id must be a positive integer",
		)
		return 0, false
	}

	return id, true
}

func newDepartmentResponse(department domain.Department) catalogResponse {
	return catalogResponse{ID: department.ID, Name: department.Name}
}

func newPositionResponse(position domain.Position) catalogResponse {
	return catalogResponse{ID: position.ID, Name: position.Name}
}
