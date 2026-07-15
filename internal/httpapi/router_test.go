package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CarambaG/employee-requests/internal/httpapi"
)

type pingerFunc func(context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

func TestHealthReturnsOKWhenDatabaseIsAvailable(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	httpapi.NewRouter(pingerFunc(func(context.Context) error {
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

	httpapi.NewRouter(pingerFunc(func(context.Context) error {
		return errors.New("database is unavailable")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"unexpected status code: got %d, want %d",
			recorder.Code,
			http.StatusServiceUnavailable,
		)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"unavailable"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
