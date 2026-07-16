package postgres

import (
	"errors"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapDatabaseError(t *testing.T) {
	if err := mapDatabaseError(pgx.ErrNoRows); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	foreignKeyError := &pgconn.PgError{
		Code:           "23503",
		ConstraintName: "employees_department_id_fkey",
	}
	if err := mapDatabaseError(foreignKeyError); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	checkError := &pgconn.PgError{
		Code:           "23514",
		ConstraintName: "requests_due_at_not_before_creation",
	}
	if err := mapDatabaseError(checkError); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
