#!/bin/sh
set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"

performance_dir=${PERFORMANCE_DIR:-/performance}

restore_index() {
    psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 <<'SQL'
CREATE INDEX IF NOT EXISTS requests_assignee_status_due_at_idx
    ON requests (assignee_id, status_id, due_at);
ANALYZE requests;
SQL
}

trap restore_index EXIT

printf '%s\n' 'Dropping the performance index for the baseline measurement...'
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c \
    'DROP INDEX IF EXISTS requests_assignee_status_due_at_idx;'

"$performance_dir/run_baseline.sh"

printf '%s\n' 'Creating the performance index for the optimized measurement...'
restore_index
trap - EXIT

"$performance_dir/run_optimized.sh"
