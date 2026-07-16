#!/bin/sh
set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"

output_dir=${OUTPUT_DIR:-/results}
performance_dir=${PERFORMANCE_DIR:-/performance}
report_file=${REPORT_FILE:-baseline.txt}
rows_file=${ROWS_FILE:-baseline_rows.csv}

mkdir -p "$output_dir"
report_path="$output_dir/$report_file"
rows_path="$output_dir/$rows_file"

case "${ASSIGNEE_ID:-}" in
    "")
        assignee_id=$(
            psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "
                SELECT r.assignee_id
                FROM requests AS r
                JOIN request_statuses AS rs ON rs.id = r.status_id
                WHERE rs.code = 'in_progress'
                  AND r.due_at < NOW()
                GROUP BY r.assignee_id
                ORDER BY COUNT(*) DESC, r.assignee_id
                LIMIT 1;
            "
        )
        ;;
    *[!0-9]*|0)
        echo "ASSIGNEE_ID must be a positive integer" >&2
        exit 1
        ;;
    *)
        assignee_id=$ASSIGNEE_ID
        ;;
esac

if [ -z "$assignee_id" ]; then
    echo "No assignee with overdue in-progress requests was found. Run the seed first." >&2
    exit 1
fi

request_count=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "SELECT COUNT(*) FROM requests;")
employee_count=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "SELECT COUNT(*) FROM employees;")
matched_rows=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "
    SELECT COUNT(*)
    FROM requests AS r
    WHERE r.assignee_id = $assignee_id
      AND r.status_id = (
          SELECT id FROM request_statuses WHERE code = 'in_progress'
      )
      AND r.due_at < NOW();
")

if [ "$matched_rows" -eq 0 ]; then
    echo "Assignee $assignee_id has no overdue in-progress requests." >&2
    exit 1
fi

optimized_index_exists=$(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c "
    SELECT TO_REGCLASS('public.requests_assignee_status_due_at_idx') IS NOT NULL;
")

if [ "$optimized_index_exists" = "t" ]; then
    echo "The optimized index already exists. Run the baseline on the commit before database optimization." >&2
    exit 1
fi

{
    echo "Employee Requests performance baseline"
    echo "Generated at (UTC): $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    echo "PostgreSQL: $(psql "$DATABASE_URL" -X -A -t -v ON_ERROR_STOP=1 -c 'SHOW server_version;')"
    echo "Employees: $employee_count"
    echo "Requests: $request_count"
    echo "Assignee ID: $assignee_id"
    echo "Matching rows: $matched_rows"
    echo
    echo "No composite performance index is present in this baseline commit."
    echo "The first EXPLAIN ANALYZE is a warm-up; use the following three execution times for comparison."
    echo
} > "$report_path"

psql "$DATABASE_URL" \
    -X \
    -v ON_ERROR_STOP=1 \
    -v assignee_id="$assignee_id" \
    -f "$performance_dir/baseline.sql" >> "$report_path"

psql "$DATABASE_URL" \
    -X \
    --csv \
    -v ON_ERROR_STOP=1 \
    -v assignee_id="$assignee_id" \
    -f "$performance_dir/query.sql" > "$rows_path"

printf '\nBaseline report: %s\n' "$report_path"
printf 'Sorted query result: %s\n\n' "$rows_path"
cat "$report_path"
