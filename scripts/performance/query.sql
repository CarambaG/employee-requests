\set ON_ERROR_STOP on

SELECT
    r.number,
    r.created_at,
    r.author_id,
    r.assignee_id,
    r.description,
    r.due_at,
    rs.code AS status,
    r.updated_at
FROM requests AS r
JOIN request_statuses AS rs ON rs.id = r.status_id
WHERE r.assignee_id = :assignee_id
  AND rs.code = 'in_progress'
  AND r.due_at < NOW()
ORDER BY r.due_at;
