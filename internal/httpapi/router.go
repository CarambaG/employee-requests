package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
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

type healthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database"`
}

func NewRouter(database Pinger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health(database))

	return mux
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

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}
