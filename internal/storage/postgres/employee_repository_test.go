package postgres

import (
	"errors"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestValidateEmployee(t *testing.T) {
	tests := []struct {
		name         string
		fullName     string
		departmentID int64
		positionID   int64
		wantError    bool
	}{
		{
			name:         "valid employee",
			fullName:     "Иванов Иван Иванович",
			departmentID: 1,
			positionID:   1,
		},
		{
			name:         "blank full name",
			fullName:     "   ",
			departmentID: 1,
			positionID:   1,
			wantError:    true,
		},
		{
			name:         "invalid department",
			fullName:     "Иванов Иван Иванович",
			departmentID: 0,
			positionID:   1,
			wantError:    true,
		},
		{
			name:         "invalid position",
			fullName:     "Иванов Иван Иванович",
			departmentID: 1,
			positionID:   -1,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmployee(tt.fullName, tt.departmentID, tt.positionID)
			if (err != nil) != tt.wantError {
				t.Fatalf("unexpected validation result: %v", err)
			}
		})
	}
}

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
}
