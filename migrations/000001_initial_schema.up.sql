CREATE TABLE departments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT departments_name_not_blank CHECK (BTRIM(name) <> ''),
    CONSTRAINT departments_name_unique UNIQUE (name)
);

CREATE TABLE positions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT positions_name_not_blank CHECK (BTRIM(name) <> ''),
    CONSTRAINT positions_name_unique UNIQUE (name)
);

CREATE TABLE employees (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    department_id BIGINT NOT NULL REFERENCES departments(id),
    position_id BIGINT NOT NULL REFERENCES positions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employees_full_name_not_blank CHECK (BTRIM(full_name) <> '')
);

CREATE INDEX employees_department_id_idx ON employees (department_id);
CREATE INDEX employees_position_id_idx ON employees (position_id);

CREATE TABLE request_statuses (
    id SMALLINT PRIMARY KEY,
    code VARCHAR(32) NOT NULL,
    name VARCHAR(64) NOT NULL,
    CONSTRAINT request_statuses_code_unique UNIQUE (code),
    CONSTRAINT request_statuses_name_unique UNIQUE (name),
    CONSTRAINT request_statuses_code_not_blank CHECK (BTRIM(code) <> ''),
    CONSTRAINT request_statuses_name_not_blank CHECK (BTRIM(name) <> '')
);

INSERT INTO request_statuses (id, code, name)
VALUES
    (1, 'new', 'Новая'),
    (2, 'in_progress', 'В работе'),
    (3, 'completed', 'Выполнена');

CREATE TABLE request_status_transitions (
    from_status_id SMALLINT NOT NULL REFERENCES request_statuses(id),
    to_status_id SMALLINT NOT NULL REFERENCES request_statuses(id),
    PRIMARY KEY (from_status_id, to_status_id),
    CONSTRAINT request_status_transition_changes_status CHECK (from_status_id <> to_status_id)
);

INSERT INTO request_status_transitions (from_status_id, to_status_id)
VALUES
    (1, 2),
    (2, 3);

CREATE TABLE requests (
    number BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    author_id BIGINT NOT NULL REFERENCES employees(id),
    assignee_id BIGINT NOT NULL REFERENCES employees(id),
    description TEXT NOT NULL,
    due_at TIMESTAMPTZ NOT NULL,
    status_id SMALLINT NOT NULL DEFAULT 1 REFERENCES request_statuses(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT requests_description_not_blank CHECK (BTRIM(description) <> ''),
    CONSTRAINT requests_due_at_not_before_creation CHECK (due_at >= created_at)
);
