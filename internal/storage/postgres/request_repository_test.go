package postgres

import (
	"strings"
	"testing"

	"github.com/CarambaG/employee-requests/internal/domain"
	requestservice "github.com/CarambaG/employee-requests/internal/request"
)

func TestBuildRequestListQuery(t *testing.T) {
	status := domain.RequestStatusInProgress
	assigneeID := int64(42)
	departmentID := int64(7)
	overdue := true

	query, arguments := buildRequestListQuery(requestservice.ListFilter{
		Status:       &status,
		AssigneeID:   &assigneeID,
		DepartmentID: &departmentID,
		Overdue:      &overdue,
		Limit:        100,
		Offset:       200,
	})

	expectedParts := []string{
		"rs.code = $1",
		"r.assignee_id = $2",
		"a.department_id = $3",
		"r.due_at < NOW()",
		"ORDER BY r.number DESC",
		"LIMIT $4 OFFSET $5",
	}
	for _, part := range expectedParts {
		if !strings.Contains(query, part) {
			t.Fatalf("query does not contain %q:\n%s", part, query)
		}
	}

	if len(arguments) != 5 {
		t.Fatalf("unexpected argument count: %d", len(arguments))
	}
	if arguments[0] != status || arguments[1] != assigneeID || arguments[2] != departmentID {
		t.Fatalf("unexpected filter arguments: %#v", arguments[:3])
	}
	if arguments[3] != 100 || arguments[4] != 200 {
		t.Fatalf("unexpected pagination arguments: %#v", arguments[3:])
	}
}

func TestBuildRequestListQueryWithoutFilters(t *testing.T) {
	query, arguments := buildRequestListQuery(requestservice.ListFilter{
		Limit:  25,
		Offset: 0,
	})

	if strings.Contains(query, " WHERE ") {
		t.Fatalf("unexpected WHERE clause:\n%s", query)
	}
	if len(arguments) != 2 || arguments[0] != 25 || arguments[1] != 0 {
		t.Fatalf("unexpected arguments: %#v", arguments)
	}
}
