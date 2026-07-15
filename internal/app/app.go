package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/CarambaG/employee-requests/internal/config"
	"github.com/CarambaG/employee-requests/internal/httpapi"
)

func Run(ctx context.Context, cfg config.Config) error {
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: httpapi.DefaultReadHeaderTimeout,
		ReadTimeout:       httpapi.DefaultReadTimeout,
		WriteTimeout:      httpapi.DefaultWriteTimeout,
		IdleTimeout:       httpapi.DefaultIdleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf(
			"starting employee-requests: env=%s address=%s",
			cfg.Environment,
			cfg.HTTPAddress,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			cfg.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		err := <-serverErrors
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}

		log.Println("employee-requests stopped")
		return nil

	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}

		return nil
	}
}
