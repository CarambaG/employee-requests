package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/CarambaG/employee-requests/internal/domain"
	requestservice "github.com/CarambaG/employee-requests/internal/request"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const requestSelect = `
	SELECT
		r.number,
		r.created_at,
		r.description,
		r.due_at,
		rs.code,
		r.updated_at,
		a.id,
		a.full_name,
		ad.id,
		ad.name,
		ap.id,
		ap.name,
		e.id,
		e.full_name,
		ed.id,
		ed.name,
		ep.id,
		ep.name
	FROM requests r
	JOIN request_statuses rs ON rs.id = r.status_id
	JOIN employees a ON a.id = r.author_id
	JOIN departments ad ON ad.id = a.department_id
	JOIN positions ap ON ap.id = a.position_id
	JOIN employees e ON e.id = r.assignee_id
	JOIN departments ed ON ed.id = e.department_id
	JOIN positions ep ON ep.id = e.position_id`

type RequestRepository struct {
	pool *pgxpool.Pool
}

var _ requestservice.Repository = (*RequestRepository)(nil)

func NewRequestRepository(pool *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{pool: pool}
}

func (r *RequestRepository) Create(
	ctx context.Context,
	params requestservice.CreateParams,
) (domain.Request, error) {
	const query = `
		WITH created_request AS (
			INSERT INTO requests (author_id, assignee_id, description, due_at)
			VALUES ($1, $2, $3, $4)
			RETURNING number
		)
	` + requestSelect + `
	JOIN created_request cr ON cr.number = r.number`

	created, err := scanRequest(r.pool.QueryRow(
		ctx,
		query,
		params.AuthorID,
		params.AssigneeID,
		params.Description,
		params.DueAt,
	))
	if err != nil {
		return domain.Request{}, mapDatabaseError(err)
	}

	return created, nil
}

func (r *RequestRepository) GetByNumber(
	ctx context.Context,
	number int64,
) (domain.Request, error) {
	query := requestSelect + ` WHERE r.number = $1`

	found, err := scanRequest(r.pool.QueryRow(ctx, query, number))
	if err != nil {
		return domain.Request{}, mapDatabaseError(err)
	}

	return found, nil
}

func (r *RequestRepository) List(
	ctx context.Context,
	filter requestservice.ListFilter,
) ([]domain.Request, error) {
	query, arguments := buildRequestListQuery(filter)

	rows, err := r.pool.Query(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	requests := make([]domain.Request, 0)
	for rows.Next() {
		found, err := scanRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		requests = append(requests, found)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate requests: %w", err)
	}

	return requests, nil
}

func (r *RequestRepository) UpdateStatus(
	ctx context.Context,
	number int64,
	next domain.RequestStatus,
) (domain.Request, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Request{}, fmt.Errorf("begin status update transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const lockQuery = `
		SELECT rs.code
		FROM requests r
		JOIN request_statuses rs ON rs.id = r.status_id
		WHERE r.number = $1
		FOR UPDATE OF r`

	var current domain.RequestStatus
	if err := tx.QueryRow(ctx, lockQuery, number).Scan(&current); err != nil {
		return domain.Request{}, mapDatabaseError(err)
	}

	request := domain.Request{Status: current}
	if err := request.ChangeStatus(next); err != nil {
		return domain.Request{}, err
	}

	const updateQuery = `
		UPDATE requests
		SET
			status_id = (SELECT id FROM request_statuses WHERE code = $2),
			updated_at = NOW()
		WHERE number = $1`

	commandTag, err := tx.Exec(ctx, updateQuery, number, next)
	if err != nil {
		return domain.Request{}, mapDatabaseError(err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.Request{}, domain.ErrNotFound
	}

	updated, err := scanRequest(tx.QueryRow(ctx, requestSelect+` WHERE r.number = $1`, number))
	if err != nil {
		return domain.Request{}, mapDatabaseError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Request{}, fmt.Errorf("commit status update transaction: %w", err)
	}

	return updated, nil
}

func (r *RequestRepository) UpdateAssignee(
	ctx context.Context,
	number int64,
	assigneeID int64,
) (domain.Request, error) {
	const query = `
		WITH updated_request AS (
			UPDATE requests
			SET
				assignee_id = $2,
				updated_at = NOW()
			WHERE number = $1
			RETURNING number
		)
	` + requestSelect + `
	JOIN updated_request ur ON ur.number = r.number`

	updated, err := scanRequest(r.pool.QueryRow(ctx, query, number, assigneeID))
	if err != nil {
		return domain.Request{}, mapDatabaseError(err)
	}

	return updated, nil
}

func buildRequestListQuery(filter requestservice.ListFilter) (string, []any) {
	conditions := make([]string, 0, 4)
	arguments := make([]any, 0, 6)

	addArgument := func(value any) string {
		arguments = append(arguments, value)
		return fmt.Sprintf("$%d", len(arguments))
	}

	if filter.Status != nil {
		conditions = append(conditions, "rs.code = "+addArgument(*filter.Status))
	}
	if filter.AssigneeID != nil {
		conditions = append(conditions, "r.assignee_id = "+addArgument(*filter.AssigneeID))
	}
	if filter.DepartmentID != nil {
		conditions = append(conditions, "a.department_id = "+addArgument(*filter.DepartmentID))
	}
	if filter.Overdue != nil {
		if *filter.Overdue {
			conditions = append(conditions, "r.due_at < NOW() AND rs.code <> 'completed'")
		} else {
			conditions = append(conditions, "(r.due_at >= NOW() OR rs.code = 'completed')")
		}
	}

	var builder strings.Builder
	builder.Grow(len(requestSelect) + 256)
	builder.WriteString(requestSelect)
	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}
	builder.WriteString(" ORDER BY r.number DESC")
	builder.WriteString(" LIMIT ")
	builder.WriteString(addArgument(filter.Limit))
	builder.WriteString(" OFFSET ")
	builder.WriteString(addArgument(filter.Offset))

	return builder.String(), arguments
}

func scanRequest(row pgx.Row) (domain.Request, error) {
	var found domain.Request

	err := row.Scan(
		&found.Number,
		&found.CreatedAt,
		&found.Description,
		&found.DueAt,
		&found.Status,
		&found.UpdatedAt,
		&found.Author.ID,
		&found.Author.FullName,
		&found.Author.Department.ID,
		&found.Author.Department.Name,
		&found.Author.Position.ID,
		&found.Author.Position.Name,
		&found.Assignee.ID,
		&found.Assignee.FullName,
		&found.Assignee.Department.ID,
		&found.Assignee.Department.Name,
		&found.Assignee.Position.ID,
		&found.Assignee.Position.Name,
	)
	if err != nil {
		return domain.Request{}, err
	}

	return found, nil
}
