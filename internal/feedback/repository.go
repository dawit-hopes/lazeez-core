package feedback

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type FeedbackRepository interface {
	UpsertOrderRating(ctx context.Context, rating OrderRating) (*OrderRating, error)
	GetOrderRatingByOrderID(ctx context.Context, orderID string) (*OrderRating, error)
	GetOrderRatingByID(ctx context.Context, id string) (*OrderRating, error)
	GetOrderRatingByIDAndBranch(ctx context.Context, id, branchID string) (*OrderRating, error)
	GetOrderRatingByIDForAdmin(ctx context.Context, id, merchantID string) (*OrderRating, error)
	ListOrderRatings(ctx context.Context, filter RatingFilter, baseCond string, baseArgs []any) (*common.PaginatedResponse[[]*OrderRating], error)

	UpsertStayRating(ctx context.Context, rating StayRating) (*StayRating, error)
	GetStayRatingBySessionKey(ctx context.Context, sessionKey string) (*StayRating, error)
	GetStayRatingByID(ctx context.Context, id string) (*StayRating, error)
	GetStayRatingByIDAndBranch(ctx context.Context, id, branchID string) (*StayRating, error)
	GetStayRatingByIDForAdmin(ctx context.Context, id, merchantID string) (*StayRating, error)
	ListStayRatings(ctx context.Context, filter RatingFilter, baseCond string, baseArgs []any) (*common.PaginatedResponse[[]*StayRating], error)
}

type feedbackRepository struct {
	join   *common.JoinDAL
	logger config.Logger
}

func NewFeedbackRepository(join *common.JoinDAL, logger config.Logger) FeedbackRepository {
	return &feedbackRepository{join: join, logger: logger}
}

const orderRatingSelect = `
SELECT r.id, r.order_id, r.branch_id, r.session_key, r.rating, r.comment, r.tags, r.phone_number,
	r.is_deleted, r.deleted_at, r.created_at, r.updated_at, o.order_number
FROM order_ratings r
INNER JOIN orders o ON o.id = r.order_id AND o.is_deleted = FALSE
WHERE r.is_deleted = FALSE`

const stayRatingSelect = `
SELECT r.id, r.session_key, r.branch_id, r.table_name, r.rating, r.comment, r.tags, r.phone_number,
	r.is_deleted, r.deleted_at, r.created_at, r.updated_at
FROM stay_ratings r
WHERE r.is_deleted = FALSE`

const upsertOrderRatingQuery = `
INSERT INTO order_ratings (id, order_id, branch_id, session_key, rating, comment, tags, phone_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (order_id) WHERE is_deleted = FALSE
DO UPDATE SET
	rating = EXCLUDED.rating,
	comment = EXCLUDED.comment,
	tags = EXCLUDED.tags,
	phone_number = EXCLUDED.phone_number,
	session_key = EXCLUDED.session_key,
	updated_at = NOW()
RETURNING id`

const upsertStayRatingQuery = `
INSERT INTO stay_ratings (id, session_key, branch_id, table_name, rating, comment, tags, phone_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (session_key) WHERE is_deleted = FALSE
DO UPDATE SET
	rating = EXCLUDED.rating,
	comment = EXCLUDED.comment,
	tags = EXCLUDED.tags,
	phone_number = EXCLUDED.phone_number,
	table_name = EXCLUDED.table_name,
	updated_at = NOW()
RETURNING id`

func (r *feedbackRepository) UpsertOrderRating(ctx context.Context, rating OrderRating) (*OrderRating, error) {
	if rating.ID == "" {
		rating.ID = common.GenerateUUID()
	}

	var returnedID string
	err := r.join.QueryRow(ctx, upsertOrderRatingQuery, []any{
		rating.ID,
		rating.OrderID,
		rating.BranchID,
		rating.SessionKey,
		rating.Rating,
		rating.Comment,
		rating.Tags,
		rating.PhoneNumber,
	}, func(row *sql.Row) error {
		return row.Scan(&returnedID)
	})
	if err != nil {
		r.logger.Error("failed to upsert order rating", "error", err)
		return nil, common.ErrInternalServerError
	}

	return r.GetOrderRatingByID(ctx, returnedID)
}

func (r *feedbackRepository) UpsertStayRating(ctx context.Context, rating StayRating) (*StayRating, error) {
	if rating.ID == "" {
		rating.ID = common.GenerateUUID()
	}

	var returnedID string
	err := r.join.QueryRow(ctx, upsertStayRatingQuery, []any{
		rating.ID,
		rating.SessionKey,
		rating.BranchID,
		rating.TableName,
		rating.Rating,
		rating.Comment,
		rating.Tags,
		rating.PhoneNumber,
	}, func(row *sql.Row) error {
		return row.Scan(&returnedID)
	})
	if err != nil {
		r.logger.Error("failed to upsert stay rating", "error", err)
		return nil, common.ErrInternalServerError
	}

	return r.GetStayRatingByID(ctx, returnedID)
}

func (r *feedbackRepository) GetOrderRatingByOrderID(ctx context.Context, orderID string) (*OrderRating, error) {
	query := orderRatingSelect + " AND r.order_id = $1"
	return r.scanOrderRating(ctx, query, []any{orderID})
}

func (r *feedbackRepository) GetOrderRatingByID(ctx context.Context, id string) (*OrderRating, error) {
	query := orderRatingSelect + " AND r.id = $1"
	return r.scanOrderRating(ctx, query, []any{id})
}

func (r *feedbackRepository) GetOrderRatingByIDAndBranch(ctx context.Context, id, branchID string) (*OrderRating, error) {
	query := orderRatingSelect + " AND r.id = $1 AND r.branch_id = $2"
	return r.scanOrderRating(ctx, query, []any{id, branchID})
}

func (r *feedbackRepository) GetOrderRatingByIDForAdmin(ctx context.Context, id, merchantID string) (*OrderRating, error) {
	query := orderRatingSelect + " AND r.id = $1"
	args := []any{id}
	if merchantID != "" {
		query += " AND r.branch_id IN (SELECT b.id FROM branches b WHERE b.merchant_id = $2 AND b.is_deleted = FALSE)"
		args = append(args, merchantID)
	}
	return r.scanOrderRating(ctx, query, args)
}

func (r *feedbackRepository) GetStayRatingBySessionKey(ctx context.Context, sessionKey string) (*StayRating, error) {
	query := stayRatingSelect + " AND r.session_key = $1"
	return r.scanStayRating(ctx, query, []any{sessionKey})
}

func (r *feedbackRepository) GetStayRatingByID(ctx context.Context, id string) (*StayRating, error) {
	query := stayRatingSelect + " AND r.id = $1"
	return r.scanStayRating(ctx, query, []any{id})
}

func (r *feedbackRepository) GetStayRatingByIDAndBranch(ctx context.Context, id, branchID string) (*StayRating, error) {
	query := stayRatingSelect + " AND r.id = $1 AND r.branch_id = $2"
	return r.scanStayRating(ctx, query, []any{id, branchID})
}

func (r *feedbackRepository) GetStayRatingByIDForAdmin(ctx context.Context, id, merchantID string) (*StayRating, error) {
	query := stayRatingSelect + " AND r.id = $1"
	args := []any{id}
	if merchantID != "" {
		query += " AND r.branch_id IN (SELECT b.id FROM branches b WHERE b.merchant_id = $2 AND b.is_deleted = FALSE)"
		args = append(args, merchantID)
	}
	return r.scanStayRating(ctx, query, args)
}

func (r *feedbackRepository) ListOrderRatings(ctx context.Context, filter RatingFilter, baseCond string, baseArgs []any) (*common.PaginatedResponse[[]*OrderRating], error) {
	return r.listOrderRatings(ctx, filter, baseCond, baseArgs)
}

func (r *feedbackRepository) ListStayRatings(ctx context.Context, filter RatingFilter, baseCond string, baseArgs []any) (*common.PaginatedResponse[[]*StayRating], error) {
	return r.listStayRatings(ctx, filter, baseCond, baseArgs)
}

func (r *feedbackRepository) listOrderRatings(ctx context.Context, filter RatingFilter, baseCond string, baseArgs []any) (*common.PaginatedResponse[[]*OrderRating], error) {
	common.NormalizeFilter(&filter.Filter)
	page, limit := filter.PageLimit()
	filterCond, args := buildOrderRatingFilterClause(filter, baseArgs)

	countQuery := "SELECT COUNT(*) FROM order_ratings r INNER JOIN orders o ON o.id = r.order_id AND o.is_deleted = FALSE WHERE r.is_deleted = FALSE" + baseCond + filterCond
	var total int64
	if err := r.join.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	}); err != nil {
		r.logger.Error("failed to count order ratings", "error", err)
		return nil, common.ErrInternalServerError
	}

	offset := (page - 1) * limit
	argNum := len(args) + 1
	query := orderRatingSelect + baseCond + filterCond + " ORDER BY r.created_at DESC LIMIT $" + strconv.Itoa(argNum) + " OFFSET $" + strconv.Itoa(argNum+1)
	listArgs := append(append([]any{}, args...), limit, offset)

	results, err := common.QueryRows(r.join, ctx, query, listArgs, func(rows *sql.Rows) (*OrderRating, error) {
		var rating OrderRating
		if err := scanOrderRatingRow(rows, &rating); err != nil {
			return nil, err
		}
		return &rating, nil
	})
	if err != nil {
		r.logger.Error("failed to list order ratings", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*OrderRating]{
		Data: results,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func (r *feedbackRepository) listStayRatings(ctx context.Context, filter RatingFilter, baseCond string, baseArgs []any) (*common.PaginatedResponse[[]*StayRating], error) {
	common.NormalizeFilter(&filter.Filter)
	page, limit := filter.PageLimit()
	filterCond, args := buildStayRatingFilterClause(filter, baseArgs)

	countQuery := "SELECT COUNT(*) FROM stay_ratings r WHERE r.is_deleted = FALSE" + baseCond + filterCond
	var total int64
	if err := r.join.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	}); err != nil {
		r.logger.Error("failed to count stay ratings", "error", err)
		return nil, common.ErrInternalServerError
	}

	offset := (page - 1) * limit
	argNum := len(args) + 1
	query := stayRatingSelect + baseCond + filterCond + " ORDER BY r.created_at DESC LIMIT $" + strconv.Itoa(argNum) + " OFFSET $" + strconv.Itoa(argNum+1)
	listArgs := append(append([]any{}, args...), limit, offset)

	results, err := common.QueryRows(r.join, ctx, query, listArgs, func(rows *sql.Rows) (*StayRating, error) {
		var rating StayRating
		if err := scanStayRatingRow(rows, &rating); err != nil {
			return nil, err
		}
		return &rating, nil
	})
	if err != nil {
		r.logger.Error("failed to list stay ratings", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*StayRating]{
		Data: results,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func (r *feedbackRepository) scanOrderRating(ctx context.Context, query string, args []any) (*OrderRating, error) {
	var rating OrderRating
	err := r.join.QueryRow(ctx, query, args, func(row *sql.Row) error {
		return scanOrderRatingRow(row, &rating)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrOrderRatingNotFound
		}
		r.logger.Error("failed to get order rating", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &rating, nil
}

func (r *feedbackRepository) scanStayRating(ctx context.Context, query string, args []any) (*StayRating, error) {
	var rating StayRating
	err := r.join.QueryRow(ctx, query, args, func(row *sql.Row) error {
		return scanStayRatingRow(row, &rating)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrStayRatingNotFound
		}
		r.logger.Error("failed to get stay rating", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &rating, nil
}

type orderRatingScanner interface {
	Scan(dest ...any) error
}

func scanOrderRatingRow(scanner orderRatingScanner, rating *OrderRating) error {
	return scanner.Scan(
		&rating.ID,
		&rating.OrderID,
		&rating.BranchID,
		&rating.SessionKey,
		&rating.Rating,
		&rating.Comment,
		&rating.Tags,
		&rating.PhoneNumber,
		&rating.IsDeleted,
		&rating.DeletedAt,
		&rating.CreatedAt,
		&rating.UpdatedAt,
		&rating.OrderNumber,
	)
}

func scanStayRatingRow(scanner orderRatingScanner, rating *StayRating) error {
	return scanner.Scan(
		&rating.ID,
		&rating.SessionKey,
		&rating.BranchID,
		&rating.TableName,
		&rating.Rating,
		&rating.Comment,
		&rating.Tags,
		&rating.PhoneNumber,
		&rating.IsDeleted,
		&rating.DeletedAt,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)
}

func orderRatingFromRequest(orderID, branchID, sessionKey string, req RatingRequest) OrderRating {
	comment := sql.NullString{}
	if req.Comment != "" {
		comment = sql.NullString{String: req.Comment, Valid: true}
	}
	return OrderRating{
		OrderID:     orderID,
		BranchID:    branchID,
		SessionKey:  sessionKey,
		Rating:      req.Rating,
		Comment:     comment,
		Tags:        tagsFromRequest(req.Tags),
		PhoneNumber: nullStringFromOptional(req.PhoneNumber),
	}
}

func stayRatingFromRequest(sessionKey, branchID string, req StayRatingRequest) StayRating {
	comment := sql.NullString{}
	if req.Comment != "" {
		comment = sql.NullString{String: req.Comment, Valid: true}
	}
	return StayRating{
		SessionKey:  sessionKey,
		BranchID:    branchID,
		TableName:   nullStringFromOptional(req.TableName),
		Rating:      req.Rating,
		Comment:     comment,
		Tags:        tagsFromRequest(req.Tags),
		PhoneNumber: nullStringFromOptional(req.PhoneNumber),
	}
}
