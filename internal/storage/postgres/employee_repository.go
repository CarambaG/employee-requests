package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/CarambaG/employee-requests/internal/employee"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmployeeRepository struct {
	pool *pgxpool.Pool
}

var _ employee.Repository = (*EmployeeRepository)(nil)

func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{pool: pool}
}

func (r *EmployeeRepository) Create(
	ctx context.Context,
	params employee.CreateParams,
) (domain.Employee, error) {
	const query = `
		WITH created_employee AS (
			INSERT INTO employees (full_name, department_id, position_id)
			VALUES ($1, $2, $3)
			RETURNING id, full_name, department_id, position_id
		)
		SELECT
			e.id,
			e.full_name,
			d.id,
			d.name,
			p.id,
			p.name
		FROM created_employee e
		JOIN departments d ON d.id = e.department_id
		JOIN positions p ON p.id = e.position_id`

	created, err := scanEmployee(r.pool.QueryRow(
		ctx,
		query,
		params.FullName,
		params.DepartmentID,
		params.PositionID,
	))
	if err != nil {
		return domain.Employee{}, mapDatabaseError(err)
	}

	return created, nil
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id int64) (domain.Employee, error) {
	const query = `
		SELECT
			e.id,
			e.full_name,
			d.id,
			d.name,
			p.id,
			p.name
		FROM employees e
		JOIN departments d ON d.id = e.department_id
		JOIN positions p ON p.id = e.position_id
		WHERE e.id = $1`

	found, err := scanEmployee(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.Employee{}, mapDatabaseError(err)
	}

	return found, nil
}

func (r *EmployeeRepository) List(ctx context.Context) ([]domain.Employee, error) {
	const query = `
		SELECT
			e.id,
			e.full_name,
			d.id,
			d.name,
			p.id,
			p.name
		FROM employees e
		JOIN departments d ON d.id = e.department_id
		JOIN positions p ON p.id = e.position_id
		ORDER BY e.id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	employees := make([]domain.Employee, 0)
	for rows.Next() {
		found, err := scanEmployee(rows)
		if err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		employees = append(employees, found)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employees: %w", err)
	}

	return employees, nil
}

func (r *EmployeeRepository) Update(
	ctx context.Context,
	id int64,
	params employee.UpdateParams,
) (domain.Employee, error) {
	const query = `
		WITH updated_employee AS (
			UPDATE employees
			SET
				full_name = $2,
				department_id = $3,
				position_id = $4,
				updated_at = NOW()
			WHERE id = $1
			RETURNING id, full_name, department_id, position_id
		)
		SELECT
			e.id,
			e.full_name,
			d.id,
			d.name,
			p.id,
			p.name
		FROM updated_employee e
		JOIN departments d ON d.id = e.department_id
		JOIN positions p ON p.id = e.position_id`

	updated, err := scanEmployee(r.pool.QueryRow(
		ctx,
		query,
		id,
		params.FullName,
		params.DepartmentID,
		params.PositionID,
	))
	if err != nil {
		return domain.Employee{}, mapDatabaseError(err)
	}

	return updated, nil
}

func (r *EmployeeRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM employees WHERE id = $1 RETURNING id`

	var deletedID int64
	if err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID); err != nil {
		return mapDatabaseError(err)
	}

	return nil
}

func scanEmployee(row pgx.Row) (domain.Employee, error) {
	var found domain.Employee

	err := row.Scan(
		&found.ID,
		&found.FullName,
		&found.Department.ID,
		&found.Department.Name,
		&found.Position.ID,
		&found.Position.Name,
	)
	if err != nil {
		return domain.Employee{}, err
	}

	return found, nil
}

func mapDatabaseError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23503", "23505":
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgError.ConstraintName)
		}
	}

	return err
}
