package common

import (
	"context"
	"database/sql"
)

// JoinDAL provides methods for executing custom queries with joins.
// It is reusable for any query that requires multiple table joins or
// complex result mapping beyond the standard DAL CRUD operations.
type JoinDAL struct {
	db *sql.DB
}

// NewJoinDAL creates a new JoinDAL instance.
func NewJoinDAL(db *sql.DB) *JoinDAL {
	return &JoinDAL{db: db}
}

// QueryRows executes a query and maps each row using the provided scan function.
// The scan function is responsible for reading the row into the desired type.
// Returns an error if the query fails or if any row scan fails.
func QueryRows[T any](j *JoinDAL, ctx context.Context, query string, args []any, scan func(*sql.Rows) (T, error)) ([]T, error) {
	rows, err := j.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// QueryRow executes a query that returns at most one row.
// The scan function is responsible for reading the row into the desired type.
// Returns sql.ErrNoRows if no row was found.
func (j *JoinDAL) QueryRow(ctx context.Context, query string, args []any, scan func(*sql.Row) error) error {
	row := j.db.QueryRowContext(ctx, query, args...)
	return scan(row)
}

// Exec executes a non-query statement (INSERT/UPDATE/DELETE or complex
// statements with CTEs) and returns the sql.Result.
func (j *JoinDAL) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return j.db.ExecContext(ctx, query, args...)
}
