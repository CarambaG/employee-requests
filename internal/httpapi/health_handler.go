package httpapi

import "net/http"

type healthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database"`
}

func health(database Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := healthResponse{
			Status:   "ok",
			Service:  "employee-requests",
			Database: "available",
		}
		statusCode := http.StatusOK

		if err := database.Ping(r.Context()); err != nil {
			response.Status = "unavailable"
			response.Database = "unavailable"
			statusCode = http.StatusServiceUnavailable
		}

		writeJSON(w, statusCode, response)
	}
}
