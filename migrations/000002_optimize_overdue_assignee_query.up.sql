CREATE INDEX requests_assignee_status_due_at_idx
    ON requests (assignee_id, status_id, due_at);
