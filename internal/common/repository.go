package common

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository[T Mappable] interface {
	Create(ctx context.Context, model T) (T, error)
	Update(ctx context.Context, model T) (T, error)
	Get(ctx context.Context, filters map[string]any) (T, error)
	List(ctx context.Context, filters map[string]any, limit, offset int) ([]T, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}

type DAL[T Mappable] struct {
	db      *sql.DB
	factory func() T
}

func NewDAL[T Mappable](db *sql.DB, factory func() T) *DAL[T] {
	return &DAL[T]{
		db:      db,
		factory: factory,
	}
}

// Get retrieves a single record by filters
func (r *DAL[T]) Get(ctx context.Context, filters map[string]any) (T, error) {
	instance := r.factory()

	// Include created_at and updated_at in SELECT
	cols := instance.Columns()
	selectCols := append(cols, "created_at", "updated_at")

	// Always filter out soft-deleted records unless explicitly requested
	if filters == nil {
		filters = make(map[string]any)
	}
	if _, exists := filters["is_deleted"]; !exists {
		filters["is_deleted"] = false
	}

	whereClause, args := r.buildWhereClause(filters, 0)
	query := fmt.Sprintf("SELECT %s FROM %s %s LIMIT 1",
		strings.Join(selectCols, ", "),
		instance.Table(),
		whereClause,
	)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(instance.Addr()...)
	if err != nil {
		return instance, err
	}
	return instance, nil
}

// List retrieves multiple records by filters
func (r *DAL[T]) List(ctx context.Context, filters map[string]any, limit, offset int) ([]T, error) {
	instance := r.factory()

	// Include created_at and updated_at in SELECT
	cols := instance.Columns()
	selectCols := append(cols, "created_at", "updated_at")

	// Always filter out soft-deleted records unless explicitly requested
	if filters == nil {
		filters = make(map[string]any)
	}
	if _, exists := filters["is_deleted"]; !exists {
		filters["is_deleted"] = false
	}

	whereClause, args := r.buildWhereClause(filters, 0)

	// Handle limit: if 0, use a large number to get all records
	if limit <= 0 {
		limit = 10000
	}

	// Always return newest records first
	orderBy := " ORDER BY created_at DESC"

	argCount := len(args)
	query := fmt.Sprintf("SELECT %s FROM %s %s%s LIMIT $%d OFFSET $%d",
		strings.Join(selectCols, ", "),
		instance.Table(),
		whereClause,
		orderBy,
		argCount+1,
		argCount+2,
	)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item := r.factory()
		if err := rows.Scan(item.Addr()...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// Create creates a new record
func (r *DAL[T]) Create(ctx context.Context, model T) (T, error) {
	cols := model.Columns()
	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Build RETURNING clause with all columns including created_at and updated_at
	returningCols := append(cols, "created_at", "updated_at")

	query := fmt.Sprintf("INSERT INTO %s (%s, updated_at, created_at) VALUES (%s, NOW(), NOW()) RETURNING %s",
		model.Table(),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(returningCols, ", "),
	)

	err := r.db.QueryRowContext(ctx, query, model.Values()...).Scan(model.Addr()...)
	return model, err
}

// Update updates a record
func (r *DAL[T]) Update(ctx context.Context, id string, model T) (T, error) {
	cols := model.Columns()
	setClauses := make([]string, len(cols))
	values := model.Values()

	for i, col := range cols {
		setClauses[i] = fmt.Sprintf("%s = $%d", col, i+2)
	}

	// Include created_at and updated_at in RETURNING
	returningCols := append(cols, "created_at", "updated_at")

	query := fmt.Sprintf("UPDATE %s SET %s, updated_at = NOW() WHERE id = $1 RETURNING %s",
		model.Table(),
		strings.Join(setClauses, ", "),
		strings.Join(returningCols, ", "),
	)

	args := append([]any{id}, values...)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(model.Addr()...)
	return model, err
}

// Delete deletes a record
func (r *DAL[T]) Delete(ctx context.Context, id string) error {
	temp := r.factory()
	query := fmt.Sprintf("UPDATE %s SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW() WHERE id = $1", temp.Table())
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Count counts the number of records
func (r *DAL[T]) Count(ctx context.Context) (int, error) {
	temp := r.factory()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE is_deleted = FALSE", temp.Table())

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *DAL[T]) buildWhereClause(filters map[string]any, startAt int) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	var clauses []string
	var args []any

	i := startAt + 1
	for col, val := range filters {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}
