BEGIN;

SET LOCAL synchronous_commit = OFF;

TRUNCATE TABLE
    requests,
    employees,
    departments,
    positions
RESTART IDENTITY CASCADE;

INSERT INTO departments (name)
SELECT FORMAT('Подразделение %s', series_number)
FROM GENERATE_SERIES(1, 20) AS series_number;

INSERT INTO positions (name)
SELECT FORMAT('Должность %s', series_number)
FROM GENERATE_SERIES(1, 50) AS series_number;

INSERT INTO employees (full_name, department_id, position_id)
SELECT
    FORMAT('Сотрудник %s', series_number),
    ((series_number - 1) % 20) + 1,
    ((series_number - 1) % 50) + 1
FROM GENERATE_SERIES(1, :employee_count::BIGINT) AS series_number;

WITH generated_requests AS (
    SELECT
        series_number,
        NOW()
            - MAKE_INTERVAL(days => (series_number % 365)::INTEGER)
            - MAKE_INTERVAL(secs => (series_number % 86400)::DOUBLE PRECISION) AS created_at
    FROM GENERATE_SERIES(1, :request_count::BIGINT) AS series_number
)
INSERT INTO requests (
    created_at,
    author_id,
    assignee_id,
    description,
    due_at,
    status_id,
    updated_at
)
SELECT
    created_at,
    ((series_number - 1) % :employee_count::BIGINT) + 1,
    (((series_number * 37) - 1) % :employee_count::BIGINT) + 1,
    FORMAT('Тестовая заявка №%s', series_number),
    created_at + MAKE_INTERVAL(days => (1 + series_number % 30)::INTEGER),
    CASE
        WHEN series_number % 10 < 3 THEN 1
        WHEN series_number % 10 < 7 THEN 2
        ELSE 3
    END,
    LEAST(
        NOW(),
        created_at + MAKE_INTERVAL(days => (series_number % 15)::INTEGER)
    )
FROM generated_requests;

ANALYZE departments;
ANALYZE positions;
ANALYZE employees;
ANALYZE requests;

COMMIT;

SELECT
    (SELECT COUNT(*) FROM departments) AS departments,
    (SELECT COUNT(*) FROM positions) AS positions,
    (SELECT COUNT(*) FROM employees) AS employees,
    (SELECT COUNT(*) FROM requests) AS requests;
