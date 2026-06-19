package feedback

import (
	"context"
	"errors"

	"lazeez-core/config"
	"lazeez-core/internal/clientsession"
	"lazeez-core/internal/common"
	order "lazeez-core/internal/order/table"
)

type FeedbackService interface {
	SubmitOrderRating(ctx context.Context, orderID, sessionKey string, req RatingRequest) (*OrderRatingDTO, error)
	GetOrderRatingClient(ctx context.Context, orderID, sessionKey string) (*OrderRatingDTO, error)
	SubmitStayRating(ctx context.Context, sessionKey string, req StayRatingRequest) (*StayRatingDTO, error)
	GetStayRatingClient(ctx context.Context, sessionKey string) (*StayRatingDTO, error)

	GetOrderRatingStaff(ctx context.Context, id string, scope ListScope) (*OrderRatingDTO, error)
	ListOrderRatingsStaff(ctx context.Context, filter RatingFilter, scope ListScope) (*common.PaginatedResponse[[]*OrderRatingDTO], error)
	GetStayRatingStaff(ctx context.Context, id string, scope ListScope) (*StayRatingDTO, error)
	ListStayRatingsStaff(ctx context.Context, filter RatingFilter, scope ListScope) (*common.PaginatedResponse[[]*StayRatingDTO], error)
}

type feedbackService struct {
	repo                 FeedbackRepository
	orderRepo            order.OrderRepository
	clientSessionService clientsession.ClientSessionService
	logger               config.Logger
}

func NewFeedbackService(
	repo FeedbackRepository,
	orderRepo order.OrderRepository,
	clientSessionService clientsession.ClientSessionService,
	logger config.Logger,
) FeedbackService {
	return &feedbackService{
		repo:                 repo,
		orderRepo:            orderRepo,
		clientSessionService: clientSessionService,
		logger:               logger,
	}
}

func (s *feedbackService) SubmitOrderRating(ctx context.Context, orderID, sessionKey string, req RatingRequest) (*OrderRatingDTO, error) {
	ord, err := s.orderRepo.GetBySessionKey(ctx, orderID, sessionKey)
	if err != nil {
		return nil, err
	}

	rating := orderRatingFromRequest(orderID, ord.BranchID, sessionKey, req)
	saved, err := s.repo.UpsertOrderRating(ctx, rating)
	if err != nil {
		return nil, err
	}

	dto := saved.ToDTO()
	return &dto, nil
}

func (s *feedbackService) GetOrderRatingClient(ctx context.Context, orderID, sessionKey string) (*OrderRatingDTO, error) {
	if _, err := s.orderRepo.GetBySessionKey(ctx, orderID, sessionKey); err != nil {
		return nil, err
	}

	rating, err := s.repo.GetOrderRatingByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, common.ErrOrderRatingNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if rating.SessionKey != sessionKey {
		return nil, nil
	}

	dto := rating.ToDTO()
	return &dto, nil
}

func (s *feedbackService) SubmitStayRating(ctx context.Context, sessionKey string, req StayRatingRequest) (*StayRatingDTO, error) {
	session, err := s.clientSessionService.GetBySessionKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	rating := stayRatingFromRequest(sessionKey, session.BranchID, req)
	saved, err := s.repo.UpsertStayRating(ctx, rating)
	if err != nil {
		return nil, err
	}

	dto := saved.ToDTO()
	return &dto, nil
}

func (s *feedbackService) GetStayRatingClient(ctx context.Context, sessionKey string) (*StayRatingDTO, error) {
	rating, err := s.repo.GetStayRatingBySessionKey(ctx, sessionKey)
	if err != nil {
		if errors.Is(err, common.ErrStayRatingNotFound) {
			return nil, nil
		}
		return nil, err
	}

	dto := rating.ToDTO()
	return &dto, nil
}

func (s *feedbackService) GetOrderRatingStaff(ctx context.Context, id string, scope ListScope) (*OrderRatingDTO, error) {
	var rating *OrderRating
	var err error

	switch {
	case scope.BranchID != "" && scope.MerchantID == "":
		rating, err = s.repo.GetOrderRatingByIDAndBranch(ctx, id, scope.BranchID)
	case scope.MerchantID != "":
		rating, err = s.repo.GetOrderRatingByIDForAdmin(ctx, id, scope.MerchantID)
	default:
		rating, err = s.repo.GetOrderRatingByID(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	if scope.BranchID != "" && scope.MerchantID != "" && rating.BranchID != scope.BranchID {
		return nil, common.ErrOrderRatingNotFound
	}

	dto := rating.ToDTO()
	return &dto, nil
}

func (s *feedbackService) ListOrderRatingsStaff(ctx context.Context, filter RatingFilter, scope ListScope) (*common.PaginatedResponse[[]*OrderRatingDTO], error) {
	baseCond, baseArgs := staffListBaseClause(scope, &filter)
	result, err := s.repo.ListOrderRatings(ctx, filter, baseCond, baseArgs)
	if err != nil {
		return nil, err
	}
	return mapOrderRatingPage(result), nil
}

func (s *feedbackService) GetStayRatingStaff(ctx context.Context, id string, scope ListScope) (*StayRatingDTO, error) {
	var rating *StayRating
	var err error

	switch {
	case scope.BranchID != "" && scope.MerchantID == "":
		rating, err = s.repo.GetStayRatingByIDAndBranch(ctx, id, scope.BranchID)
	case scope.MerchantID != "":
		rating, err = s.repo.GetStayRatingByIDForAdmin(ctx, id, scope.MerchantID)
	default:
		rating, err = s.repo.GetStayRatingByID(ctx, id)
	}
	if err != nil {
		return nil, err
	}
	if scope.BranchID != "" && scope.MerchantID != "" && rating.BranchID != scope.BranchID {
		return nil, common.ErrStayRatingNotFound
	}

	dto := rating.ToDTO()
	return &dto, nil
}

func (s *feedbackService) ListStayRatingsStaff(ctx context.Context, filter RatingFilter, scope ListScope) (*common.PaginatedResponse[[]*StayRatingDTO], error) {
	baseCond, baseArgs := staffListBaseClause(scope, &filter)
	result, err := s.repo.ListStayRatings(ctx, filter, baseCond, baseArgs)
	if err != nil {
		return nil, err
	}
	return mapStayRatingPage(result), nil
}

func staffListBaseClause(scope ListScope, filter *RatingFilter) (string, []any) {
	if scope.MerchantID != "" {
		filter.MerchantID = scope.MerchantID
		if scope.BranchID != "" {
			filter.BranchID = scope.BranchID
		}
		return " ", []any{}
	}
	if scope.BranchID != "" {
		return " AND r.branch_id = $1 ", []any{scope.BranchID}
	}
	return " ", []any{}
}

func mapOrderRatingPage(result *common.PaginatedResponse[[]*OrderRating]) *common.PaginatedResponse[[]*OrderRatingDTO] {
	dtos := make([]*OrderRatingDTO, 0, len(result.Data))
	for _, rating := range result.Data {
		dto := rating.ToDTO()
		dtos = append(dtos, &dto)
	}
	return &common.PaginatedResponse[[]*OrderRatingDTO]{
		Data: dtos,
		Meta: result.Meta,
	}
}

func mapStayRatingPage(result *common.PaginatedResponse[[]*StayRating]) *common.PaginatedResponse[[]*StayRatingDTO] {
	dtos := make([]*StayRatingDTO, 0, len(result.Data))
	for _, rating := range result.Data {
		dto := rating.ToDTO()
		dtos = append(dtos, &dto)
	}
	return &common.PaginatedResponse[[]*StayRatingDTO]{
		Data: dtos,
		Meta: result.Meta,
	}
}
