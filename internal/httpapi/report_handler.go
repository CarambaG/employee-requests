package httpapi

import (
	"net/http"

	"github.com/CarambaG/employee-requests/internal/report"
)

type reportSummaryResponse struct {
	RequestsByStatus    []reportStatusCountResponse   `json:"requests_by_status"`
	OverdueRequests     int64                         `json:"overdue_requests"`
	CompletedByAssignee []reportAssigneeCountResponse `json:"completed_by_assignee"`
}

type reportStatusCountResponse struct {
	Status requestStatusView `json:"status"`
	Count  int64             `json:"count"`
}

type reportAssigneeCountResponse struct {
	Assignee          reportAssigneeResponse `json:"assignee"`
	CompletedRequests int64                  `json:"completed_requests"`
}

type reportAssigneeResponse struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
}

func getReportSummary(service ReportService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := service.GetSummary(r.Context())
		if err != nil {
			writeServiceError(w, err, "report")
			return
		}

		writeJSON(w, http.StatusOK, newReportSummaryResponse(summary))
	}
}

func newReportSummaryResponse(summary report.Summary) reportSummaryResponse {
	statusCounts := make([]reportStatusCountResponse, 0, len(summary.RequestsByStatus))
	for _, item := range summary.RequestsByStatus {
		statusCounts = append(statusCounts, reportStatusCountResponse{
			Status: requestStatusView{
				Code: item.Status,
				Name: requestStatusName(item.Status),
			},
			Count: item.Count,
		})
	}

	assigneeCounts := make([]reportAssigneeCountResponse, 0, len(summary.CompletedByAssignee))
	for _, item := range summary.CompletedByAssignee {
		assigneeCounts = append(assigneeCounts, reportAssigneeCountResponse{
			Assignee: reportAssigneeResponse{
				ID:       item.AssigneeID,
				FullName: item.AssigneeName,
			},
			CompletedRequests: item.Count,
		})
	}

	return reportSummaryResponse{
		RequestsByStatus:    statusCounts,
		OverdueRequests:     summary.OverdueRequests,
		CompletedByAssignee: assigneeCounts,
	}
}
