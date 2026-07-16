package postgres

import (
	"context"
	"fmt"

	"github.com/CarambaG/employee-requests/internal/catalog"
	"github.com/CarambaG/employee-requests/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogRepository struct {
	pool *pgxpool.Pool
}

var _ catalog.Repository = (*CatalogRepository)(nil)

func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) CreateDepartment(
	ctx context.Context,
	name string,
) (domain.Department, error) {
	const query = `
		INSERT INTO departments (name)
		VALUES ($1)
		RETURNING id, name`

	var created domain.Department
	if err := r.pool.QueryRow(ctx, query, name).Scan(&created.ID, &created.Name); err != nil {
		return domain.Department{}, mapDatabaseError(err)
	}

	return created, nil
}

func (r *CatalogRepository) GetDepartmentByID(
	ctx context.Context,
	id int64,
) (domain.Department, error) {
	const query = `SELECT id, name FROM departments WHERE id = $1`

	var found domain.Department
	if err := r.pool.QueryRow(ctx, query, id).Scan(&found.ID, &found.Name); err != nil {
		return domain.Department{}, mapDatabaseError(err)
	}

	return found, nil
}

func (r *CatalogRepository) ListDepartments(ctx context.Context) ([]domain.Department, error) {
	const query = `SELECT id, name FROM departments ORDER BY name, id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()

	departments := make([]domain.Department, 0)
	for rows.Next() {
		var department domain.Department
		if err := rows.Scan(&department.ID, &department.Name); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		departments = append(departments, department)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate departments: %w", err)
	}

	return departments, nil
}

func (r *CatalogRepository) UpdateDepartment(
	ctx context.Context,
	id int64,
	name string,
) (domain.Department, error) {
	const query = `
		UPDATE departments
		SET name = $2
		WHERE id = $1
		RETURNING id, name`

	var updated domain.Department
	if err := r.pool.QueryRow(ctx, query, id, name).Scan(&updated.ID, &updated.Name); err != nil {
		return domain.Department{}, mapDatabaseError(err)
	}

	return updated, nil
}

func (r *CatalogRepository) DeleteDepartment(ctx context.Context, id int64) error {
	const query = `DELETE FROM departments WHERE id = $1 RETURNING id`

	var deletedID int64
	if err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID); err != nil {
		return mapDatabaseError(err)
	}

	return nil
}

func (r *CatalogRepository) CreatePosition(
	ctx context.Context,
	name string,
) (domain.Position, error) {
	const query = `
		INSERT INTO positions (name)
		VALUES ($1)
		RETURNING id, name`

	var created domain.Position
	if err := r.pool.QueryRow(ctx, query, name).Scan(&created.ID, &created.Name); err != nil {
		return domain.Position{}, mapDatabaseError(err)
	}

	return created, nil
}

func (r *CatalogRepository) GetPositionByID(
	ctx context.Context,
	id int64,
) (domain.Position, error) {
	const query = `SELECT id, name FROM positions WHERE id = $1`

	var found domain.Position
	if err := r.pool.QueryRow(ctx, query, id).Scan(&found.ID, &found.Name); err != nil {
		return domain.Position{}, mapDatabaseError(err)
	}

	return found, nil
}

func (r *CatalogRepository) ListPositions(ctx context.Context) ([]domain.Position, error) {
	const query = `SELECT id, name FROM positions ORDER BY name, id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	defer rows.Close()

	positions := make([]domain.Position, 0)
	for rows.Next() {
		var position domain.Position
		if err := rows.Scan(&position.ID, &position.Name); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		positions = append(positions, position)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}

	return positions, nil
}

func (r *CatalogRepository) UpdatePosition(
	ctx context.Context,
	id int64,
	name string,
) (domain.Position, error) {
	const query = `
		UPDATE positions
		SET name = $2
		WHERE id = $1
		RETURNING id, name`

	var updated domain.Position
	if err := r.pool.QueryRow(ctx, query, id, name).Scan(&updated.ID, &updated.Name); err != nil {
		return domain.Position{}, mapDatabaseError(err)
	}

	return updated, nil
}

func (r *CatalogRepository) DeletePosition(ctx context.Context, id int64) error {
	const query = `DELETE FROM positions WHERE id = $1 RETURNING id`

	var deletedID int64
	if err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID); err != nil {
		return mapDatabaseError(err)
	}

	return nil
}
