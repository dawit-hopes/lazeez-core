package promotion

import (
	"context"
	"database/sql"
	"errors"

	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type PromotionRepository interface {
	Create(ctx context.Context, promotion Promotion) error
	Get(ctx context.Context, id string) (Promotion, error)
	GetByMerchant(ctx context.Context, id, merchantID string) (Promotion, error)
	Update(ctx context.Context, promotion Promotion) error
	Delete(ctx context.Context, id, merchantID string) error
	UnDelete(ctx context.Context, id, merchantID string) error
	ListByMerchant(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*Promotion], error)
	ListActivePublicByMerchant(ctx context.Context, merchantID string) ([]*Promotion, error)
	ListActivePublicByReference(ctx context.Context, reference, referenceType string) ([]*Promotion, error)
}

type promotionRepository struct {
	dal    *common.DAL[*Promotion]
	join   *common.JoinDAL
	logger config.Logger
}

func NewPromotionRepository(dal *common.DAL[*Promotion], join *common.JoinDAL, logger config.Logger) PromotionRepository {
	return &promotionRepository{dal: dal, join: join, logger: logger}
}

func (r *promotionRepository) Create(ctx context.Context, promotion Promotion) error {
	_, err := r.dal.Create(ctx, &promotion)
	if err != nil {
		r.logger.Error("failed to create promotion", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *promotionRepository) Get(ctx context.Context, id string) (Promotion, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Promotion{}, common.ErrPromotionNotFound
		}
		r.logger.Error("failed to get promotion", "error", err)
		return Promotion{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *promotionRepository) GetByMerchant(ctx context.Context, id, merchantID string) (Promotion, error) {
	filter := map[string]any{"id": id, "merchant_id": merchantID}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Promotion{}, common.ErrPromotionNotFound
		}
		r.logger.Error("failed to get promotion by merchant", "error", err)
		return Promotion{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *promotionRepository) Update(ctx context.Context, promotion Promotion) error {
	filter := map[string]any{"id": promotion.ID}
	updates := map[string]any{
		"merchant_id":   promotion.MerchantID,
		"title":         promotion.Title,
		"description":   promotion.Description,
		"banner_image":  promotion.BannerImage,
		"is_active":     promotion.IsActive,
		"start_date":    promotion.StartDate,
		"end_date":      promotion.EndDate,
		"deleted_at":    promotion.DeletedAt,
		"is_deleted":    promotion.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrPromotionNotFound
		}
		r.logger.Error("failed to update promotion", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *promotionRepository) Delete(ctx context.Context, id, merchantID string) error {
	filter := map[string]any{"id": id, "merchant_id": merchantID}
	updates := map[string]any{"is_deleted": true}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrPromotionNotFound
		}
		r.logger.Error("failed to delete promotion", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *promotionRepository) UnDelete(ctx context.Context, id, merchantID string) error {
	filter := map[string]any{"id": id, "merchant_id": merchantID}
	updates := map[string]any{"is_deleted": false, "deleted_at": nil}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrPromotionNotFound
		}
		r.logger.Error("failed to undelete promotion", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *promotionRepository) ListByMerchant(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*Promotion], error) {
	common.NormalizeFilter(&filter)
	filters := map[string]any{"merchant_id": merchantID}
	if filter.Search != "" {
		filters["title"] = common.ILike(filter.Search)
	}

	page, limit := filter.PageLimit()
	results, err := r.dal.List(ctx, filters, page, limit)
	if err != nil {
		r.logger.Error("failed to list promotions", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*Promotion]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), page, limit),
	}, nil
}

const listActivePublicByMerchantQuery = `
SELECT id, merchant_id, title, description, banner_image, is_active, start_date, end_date,
	deleted_at, is_deleted, created_at, updated_at
FROM promotions
WHERE merchant_id = $1
	AND is_deleted = FALSE
	AND is_active = TRUE
	AND start_date <= CURRENT_DATE
	AND end_date >= CURRENT_DATE
ORDER BY start_date DESC`

func (r *promotionRepository) ListActivePublicByMerchant(ctx context.Context, merchantID string) ([]*Promotion, error) {
	rows, err := common.QueryRows(r.join, ctx, listActivePublicByMerchantQuery, []any{merchantID}, func(rows *sql.Rows) (*Promotion, error) {
		var promo Promotion
		if err := rows.Scan(promo.Addr()...); err != nil {
			return nil, err
		}
		return &promo, nil
	})
	if err != nil {
		r.logger.Error("failed to list active public promotions", "error", err)
		return nil, common.ErrInternalServerError
	}
	if rows == nil {
		return []*Promotion{}, nil
	}
	return rows, nil
}

const listActivePublicByTableReferenceQuery = `
SELECT p.id, p.merchant_id, p.title, p.description, p.banner_image, p.is_active, p.start_date, p.end_date,
	p.deleted_at, p.is_deleted, p.created_at, p.updated_at
FROM tables t
INNER JOIN branches b ON b.id = t.branch_id AND b.is_deleted = FALSE
INNER JOIN promotions p ON p.merchant_id = b.merchant_id
WHERE t.reference = $1
	AND t.is_deleted = FALSE
	AND p.is_deleted = FALSE
	AND p.is_active = TRUE
	AND p.start_date <= CURRENT_DATE
	AND p.end_date >= CURRENT_DATE
ORDER BY p.start_date DESC`

const listActivePublicByRoomReferenceQuery = `
SELECT p.id, p.merchant_id, p.title, p.description, p.banner_image, p.is_active, p.start_date, p.end_date,
	p.deleted_at, p.is_deleted, p.created_at, p.updated_at
FROM single_rooms sr
INNER JOIN branches b ON b.id = sr.branch_id AND b.is_deleted = FALSE
INNER JOIN promotions p ON p.merchant_id = b.merchant_id
WHERE sr.reference = $1
	AND sr.is_deleted = FALSE
	AND p.is_deleted = FALSE
	AND p.is_active = TRUE
	AND p.start_date <= CURRENT_DATE
	AND p.end_date >= CURRENT_DATE
ORDER BY p.start_date DESC`

func (r *promotionRepository) ListActivePublicByReference(ctx context.Context, reference, referenceType string) ([]*Promotion, error) {
	query := listActivePublicByTableReferenceQuery
	if referenceType == "room" {
		query = listActivePublicByRoomReferenceQuery
	}

	rows, err := common.QueryRows(r.join, ctx, query, []any{reference}, func(rows *sql.Rows) (*Promotion, error) {
		var promo Promotion
		if err := rows.Scan(promo.Addr()...); err != nil {
			return nil, err
		}
		return &promo, nil
	})
	if err != nil {
		r.logger.Error("failed to list active public promotions by reference", "reference_type", referenceType, "error", err)
		return nil, common.ErrInternalServerError
	}
	if rows == nil {
		return []*Promotion{}, nil
	}
	return rows, nil
}
