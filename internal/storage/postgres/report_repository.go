package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CarambaG/employee-requests/internal/report"
	"github.com/jackc/pgx/v5/pgxpool"
)

const reportSummaryQuery = `
	WITH status_counts AS (
		SELECT
			rs.id,
			rs.code AS status,
			COUNT(r.number)::BIGINT AS count
		FROM request_statuses rs
		LEFT JOIN requests r ON r.status_id = rs.id
		GROUP BY rs.id, rs.code
	),
	overdue_count AS (
		SELECT COUNT(*)::BIGINT AS count
		FROM requests r
		JOIN request_statuses rs ON rs.id = r.status_id
		WHERE r.due_at < NOW()
		  AND rs.code <> 'completed'
	),
	completed_by_assignee AS (
		SELECT
			e.id AS assignee_id,
			e.full_name AS assignee_name,
			COUNT(r.number)::BIGINT AS count
		FROM employees e
		LEFT JOIN requests r
			ON r.assignee_id = e.id
			AND r.status_id = (
				SELECT id
				FROM request_statuses
				WHERE code = 'completed'
			)
		GROUP BY e.id, e.full_name
	)
	SELECT
		COALESCE(
			(
				SELECT jsonb_agg(
					jsonb_build_object('status', status, 'count', count)
					ORDER BY id
				)
				FROM status_counts
			),
			'[]'::jsonb
		),
		(SELECT count FROM overdue_count),
		COALESCE(
			(
				SELECT jsonb_agg(
					jsonb_build_object(
						'assignee_id', assignee_id,
						'assignee_name', assignee_name,
						'count', count
					)
					ORDER BY count DESC, assignee_name, assignee_id
				)
				FROM completed_by_assignee
			),
			'[]'::jsonb
		)`

type ReportRepository struct {
	pool *pgxpool.Pool
}

var _ report.Repository = (*ReportRepository)(nil)

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

func (r *ReportRepository) GetSummary(ctx context.Context) (report.Summary, error) {
	var (
		statusCountsJSON   []byte
		assigneeCountsJSON []byte
		summary            report.Summary
	)

	if err := r.pool.QueryRow(ctx, reportSummaryQuery).Scan(
		&statusCountsJSON,
		&summary.OverdueRequests,
		&assigneeCountsJSON,
	); err != nil {
		return report.Summary{}, fmt.Errorf("query request report: %w", err)
	}

	if err := json.Unmarshal(statusCountsJSON, &summary.RequestsByStatus); err != nil {
		return report.Summary{}, fmt.Errorf("decode request status counts: %w", err)
	}
	if err := json.Unmarshal(assigneeCountsJSON, &summary.CompletedByAssignee); err != nil {
		return report.Summary{}, fmt.Errorf("decode completed request counts: %w", err)
	}

	return summary, nil
}
