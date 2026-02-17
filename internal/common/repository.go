package common

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository[T Mappable] interface {
	Create(ctx context.Context, model T) (T, error)
	Update(ctx context.Context, filters map[string]any, updates map[string]any) error
	Get(ctx context.Context, filters map[string]any) (T, error)
	List(ctx context.Context, filters map[string]any, limit, offset int) ([]T, error)
	Delete(ctx context.Context, id string) error
	DeleteByFilters(ctx context.Context, filters map[string]any) error
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

	cols := instance.Columns()
	selectCols := append(cols, "created_at", "updated_at")

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
func (r *DAL[T]) List(ctx context.Context, filters map[string]any, page, limit int) ([]T, error) {
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

	offset := (page - 1) * limit

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

// ListIncludeDeleted retrieves multiple records by filters without automatically
// filtering out soft-deleted records. Callers must specify any is_deleted
// constraints explicitly in filters if desired.
func (r *DAL[T]) ListIncludeDeleted(ctx context.Context, filters map[string]any, limit, offset int) ([]T, error) {
	instance := r.factory()

	// Include created_at and updated_at in SELECT
	cols := instance.Columns()
	selectCols := append(cols, "created_at", "updated_at")

	if filters == nil {
		filters = make(map[string]any)
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

func (r *DAL[T]) Update(ctx context.Context, filters map[string]any, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	instance := r.factory()

	setClauses := make([]string, 0, len(updates))
	values := make([]any, 0, len(updates))
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	whereClause, filterArgs := r.buildWhereClause(filters, len(updates))

	query := fmt.Sprintf("UPDATE %s SET %s, updated_at = NOW() %s",
		instance.Table(),
		strings.Join(setClauses, ", "),
		whereClause,
	)

	args := append(values, filterArgs...)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deletes a record
func (r *DAL[T]) Delete(ctx context.Context, id string) error {
	temp := r.factory()
	query := fmt.Sprintf("UPDATE %s SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW() WHERE id = $1", temp.Table())
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// SoftDeleteWhere performs a set-based soft delete on the table for T using
// an arbitrary WHERE clause. This lets you delete related records in bulk,
// e.g. all users whose branch belongs to a given merchant.
//
// Example:
//
//	err := userDAL.SoftDeleteWhere(
//	    ctx,
//	    "branch_id IN (SELECT id FROM branches WHERE merchant_id = $1 AND is_deleted = FALSE)",
//	    merchantID,
//	)
func (r *DAL[T]) SoftDeleteWhere(ctx context.Context, where string, args ...any) error {
	where = strings.TrimSpace(where)
	if where == "" {
		// Avoid accidentally deleting every row
		return nil
	}

	temp := r.factory()
	query := fmt.Sprintf(
		"UPDATE %s SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW() WHERE %s",
		temp.Table(),
		where,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Count counts the number of records
func (r *DAL[T]) Count(ctx context.Context) (int, error) {
	temp := r.factory()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE is_deleted = FALSE", temp.Table())

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// ILike is a filter value that generates SQL "col ILIKE $n ESCAPE '\'" for pattern matching.
// The value is escaped for LIKE special chars (% and _) and wrapped in % for "contains" search.
type ILike string

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

// ILikePattern returns a pattern suitable for ILIKE $n ESCAPE '\' (contains search).
// Use this when building raw SQL that uses ILIKE for search.
func ILikePattern(s string) string {
	return "%" + escapeLike(s) + "%"
}

func (r *DAL[T]) buildWhereClause(filters map[string]any, startAt int) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	var clauses []string
	var args []any

	i := startAt + 1
	for col, val := range filters {
		switch v := val.(type) {
		case ILike:
			pattern := "%" + escapeLike(string(v)) + "%"
			clauses = append(clauses, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", col, i))
			args = append(args, pattern)
		default:
			clauses = append(clauses, fmt.Sprintf("%s = $%d", col, i))
			args = append(args, val)
		}
		i++
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}

func (r *DAL[T]) DeleteByFilters(ctx context.Context, filters map[string]any) error {
	if len(filters) == 0 {
		return nil
	}

	instance := r.factory()
	whereClause, args := r.buildWhereClause(filters, 0)
	query := fmt.Sprintf(
		"UPDATE %s SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW() %s",
		instance.Table(),
		whereClause,
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
