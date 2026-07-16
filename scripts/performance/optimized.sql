\set ON_ERROR_STOP on

\echo 'Warm-up run (not used as the final measurement)'
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    r.number,
    r.created_at,
    r.author_id,
    r.assignee_id,
    r.description,
    r.due_at,
    r.status_id,
    r.updated_at
FROM requests AS r
WHERE r.assignee_id = :assignee_id
  AND r.status_id = (
      SELECT id
      FROM request_statuses
      WHERE code = 'in_progress'
  )
  AND r.due_at < NOW()
ORDER BY r.due_at;

\echo 'Measured run 1 of 3'
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    r.number,
    r.created_at,
    r.author_id,
    r.assignee_id,
    r.description,
    r.due_at,
    r.status_id,
    r.updated_at
FROM requests AS r
WHERE r.assignee_id = :assignee_id
  AND r.status_id = (
      SELECT id
      FROM request_statuses
      WHERE code = 'in_progress'
  )
  AND r.due_at < NOW()
ORDER BY r.due_at;

\echo 'Measured run 2 of 3'
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    r.number,
    r.created_at,
    r.author_id,
    r.assignee_id,
    r.description,
    r.due_at,
    r.status_id,
    r.updated_at
FROM requests AS r
WHERE r.assignee_id = :assignee_id
  AND r.status_id = (
      SELECT id
      FROM request_statuses
      WHERE code = 'in_progress'
  )
  AND r.due_at < NOW()
ORDER BY r.due_at;

\echo 'Measured run 3 of 3'
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT
    r.number,
    r.created_at,
    r.author_id,
    r.assignee_id,
    r.description,
    r.due_at,
    r.status_id,
    r.updated_at
FROM requests AS r
WHERE r.assignee_id = :assignee_id
  AND r.status_id = (
      SELECT id
      FROM request_statuses
      WHERE code = 'in_progress'
  )
  AND r.due_at < NOW()
ORDER BY r.due_at;
