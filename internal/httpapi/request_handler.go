package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/CarambaG/employee-requests/internal/domain"
	requestservice "github.com/CarambaG/employee-requests/internal/request"
)

type requestCreateRequest struct {
	AuthorID    int64     `json:"author_id"`
	AssigneeID  int64     `json:"assignee_id"`
	Description string    `json:"description"`
	DueAt       time.Time `json:"due_at"`
}

type requestResponse struct {
	Number      int64             `json:"number"`
	CreatedAt   time.Time         `json:"created_at"`
	Author      employeeResponse  `json:"author"`
	Assignee    employeeResponse  `json:"assignee"`
	Description string            `json:"description"`
	DueAt       time.Time         `json:"due_at"`
	Status      requestStatusView `json:"status"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type requestStatusView struct {
	Code domain.RequestStatus `json:"code"`
	Name string               `json:"name"`
}

func createRequest(service RequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input requestCreateRequest
		if err := decodeJSON(w, r, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		created, err := service.Create(r.Context(), requestservice.CreateParams{
			AuthorID:    input.AuthorID,
			AssigneeID:  input.AssigneeID,
			Description: input.Description,
			DueAt:       input.DueAt,
		})
		if err != nil {
			writeServiceError(w, err, "request")
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/v1/requests/%d", created.Number))
		writeJSON(w, http.StatusCreated, newRequestResponse(created))
	}
}

func getRequest(service RequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		number, ok := requestNumber(w, r)
		if !ok {
			return
		}

		found, err := service.GetByNumber(r.Context(), number)
		if err != nil {
			writeServiceError(w, err, "request")
			return
		}

		writeJSON(w, http.StatusOK, newRequestResponse(found))
	}
}

func listRequests(service RequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, err := requestListFilter(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
			return
		}

		requests, err := service.List(r.Context(), filter)
		if err != nil {
			writeServiceError(w, err, "request")
			return
		}

		response := make([]requestResponse, 0, len(requests))
		for _, found := range requests {
			response = append(response, newRequestResponse(found))
		}

		writeJSON(w, http.StatusOK, response)
	}
}

type requestStatusUpdateRequest struct {
	Status domain.RequestStatus `json:"status"`
}

type requestAssigneeUpdateRequest struct {
	AssigneeID int64 `json:"assignee_id"`
}

func updateRequestStatus(service RequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		number, ok := requestNumber(w, r)
		if !ok {
			return
		}

		var input requestStatusUpdateRequest
		if err := decodeJSON(w, r, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		updated, err := service.UpdateStatus(r.Context(), number, input.Status)
		if err != nil {
			writeServiceError(w, err, "request")
			return
		}

		writeJSON(w, http.StatusOK, newRequestResponse(updated))
	}
}

func updateRequestAssignee(service RequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		number, ok := requestNumber(w, r)
		if !ok {
			return
		}

		var input requestAssigneeUpdateRequest
		if err := decodeJSON(w, r, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
			return
		}

		updated, err := service.UpdateAssignee(r.Context(), number, input.AssigneeID)
		if err != nil {
			writeServiceError(w, err, "request")
			return
		}

		writeJSON(w, http.StatusOK, newRequestResponse(updated))
	}
}

func requestNumber(w http.ResponseWriter, r *http.Request) (int64, bool) {
	number, err := strconv.ParseInt(r.PathValue("number"), 10, 64)
	if err != nil || number <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_number",
			"request number must be a positive integer",
		)
		return 0, false
	}

	return number, true
}

func requestListFilter(r *http.Request) (requestservice.ListFilter, error) {
	query := r.URL.Query()
	filter := requestservice.ListFilter{Limit: requestservice.DefaultListLimit}

	if value := query.Get("status"); value != "" {
		status := domain.RequestStatus(value)
		filter.Status = &status
	}

	if value := query.Get("assignee_id"); value != "" {
		id, err := parsePositiveInt64(value, "assignee_id")
		if err != nil {
			return requestservice.ListFilter{}, err
		}
		filter.AssigneeID = &id
	}

	if value := query.Get("department_id"); value != "" {
		id, err := parsePositiveInt64(value, "department_id")
		if err != nil {
			return requestservice.ListFilter{}, err
		}
		filter.DepartmentID = &id
	}

	if value := query.Get("overdue"); value != "" {
		overdue, err := strconv.ParseBool(value)
		if err != nil {
			return requestservice.ListFilter{}, fmt.Errorf("overdue must be true or false")
		}
		filter.Overdue = &overdue
	}

	if value := query.Get("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil {
			return requestservice.ListFilter{}, fmt.Errorf("limit must be an integer")
		}
		filter.Limit = limit
	}

	if value := query.Get("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil {
			return requestservice.ListFilter{}, fmt.Errorf("offset must be an integer")
		}
		filter.Offset = offset
	}

	return filter, nil
}

func parsePositiveInt64(value, field string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", field)
	}

	return id, nil
}

func newRequestResponse(found domain.Request) requestResponse {
	return requestResponse{
		Number:      found.Number,
		CreatedAt:   found.CreatedAt,
		Author:      newEmployeeResponse(found.Author),
		Assignee:    newEmployeeResponse(found.Assignee),
		Description: found.Description,
		DueAt:       found.DueAt,
		Status: requestStatusView{
			Code: found.Status,
			Name: requestStatusName(found.Status),
		},
		UpdatedAt: found.UpdatedAt,
	}
}

func requestStatusName(status domain.RequestStatus) string {
	switch status {
	case domain.RequestStatusNew:
		return "Новая"
	case domain.RequestStatusInProgress:
		return "В работе"
	case domain.RequestStatusCompleted:
		return "Выполнена"
	default:
		return string(status)
	}
}
