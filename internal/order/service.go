package order

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	item "lazeez-core/internal/order/Item"
	"time"

	"golang.org/x/sync/errgroup"
)

type OrderService interface {
	// Client: create order (no auth), get/list by sessionKey
	CreateClient(ctx context.Context, order OrderInput) (*OrderDTO, error)
	GetClient(ctx context.Context, id string, sessionKey string) (*OrderDTO, error)
	ListClient(ctx context.Context, filter OrderFilter, sessionKey string) (*common.PaginatedResponse[[]*OrderDTO], error)

	// Branch: get/list/update orders for branch (branch_id from token)
	GetBranch(ctx context.Context, id string, branchID string) (*OrderDTO, error)
	ListBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error)
	UpdateBranch(ctx context.Context, id string, input OrderUpdateInput, branchID string) error

	// Admin: get/list/update all orders
	GetAdmin(ctx context.Context, id string) (*OrderDTO, error)
	ListAdmin(ctx context.Context, filter OrderFilter) (*common.PaginatedResponse[[]*OrderDTO], error)
	UpdateAdmin(ctx context.Context, id string, input OrderUpdateInput) error
}

type orderService struct {
	orderRepository  OrderRepository
	orderItemService item.OrderItemService
	logger           config.Logger
}

func NewOrderService(orderRepository OrderRepository, orderItemService item.OrderItemService, logger config.Logger) OrderService {
	return &orderService{
		orderRepository:  orderRepository,
		orderItemService: orderItemService,
		logger:           logger,
	}
}

func (s *orderService) CreateClient(ctx context.Context, order OrderInput) (*OrderDTO, error) {
	// Validate total matches sum of order items
	itemsTotal := 0.0
	for _, oi := range order.OrderItems {
		itemsTotal += oi.Total
	}
	const epsilon = 0.01 // allow small floating point variance
	if itemsTotal < order.Total-epsilon || itemsTotal > order.Total+epsilon {
		s.logger.Error("order total mismatch", "order total", order.Total, "items total", itemsTotal)
		return nil, common.ErrOrderTotalMismatch
	}

	paymentMethod := order.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "other"
	}

	branchID, err := s.resolveBranchID(ctx, order.BranchID, order.SessionKey)
	if err != nil {
		return nil, err
	}

	orderNumber, err := s.assignOrderNumber(ctx, branchID)
	if err != nil {
		return nil, err
	}

	orderModel := Order{
		OrderNumber:   orderNumber,
		TableNumber:   order.TableNumber,
		BranchID:      branchID,
		SessionKey:    order.SessionKey,
		Total:         order.Total,
		PaymentMethod: paymentMethod,
		OrderStatus:   "pending",
		PaymentStatus: "pending",
		PaymentDate:   time.Now(),
	}
	orderModel.ID = common.GenerateUUID()
	created, err := s.orderRepository.Create(ctx, orderModel)
	if err != nil {
		s.logger.Error("failed to create order", "error", err)
		return nil, err
	}
	// Create order items with order_id (order must exist first)
	g, orderItemCtx := errgroup.WithContext(ctx)
	for _, oi := range order.OrderItems {
		oi := oi
		oi.OrderID = created.ID
		oi.BranchID = branchID
		g.Go(func() error {
			return s.orderItemService.Create(orderItemCtx, oi)
		})
	}
	if err := g.Wait(); err != nil {
		s.logger.Error("failed to create order items", "error", err)
		if delErr := s.orderRepository.Delete(ctx, created.ID); delErr != nil {
			s.logger.Error("failed to rollback order after item create failure", "order_id", created.ID, "error", delErr)
		}
		return nil, err
	}
	return s.orderRepository.GetBySessionKey(ctx, created.ID, order.SessionKey)
}

func (s *orderService) resolveBranchID(ctx context.Context, branchID, sessionKey string) (string, error) {
	if branchID != "" {
		return branchID, nil
	}
	if sessionKey == "" {
		s.logger.Error("branch id missing and no session key to resolve table")
		return "", common.ErrInvalidRequest
	}
	return s.orderRepository.GetBranchIDByReference(ctx, sessionKey)
}

func (s *orderService) GetClient(ctx context.Context, id string, sessionKey string) (*OrderDTO, error) {
	return s.orderRepository.GetBySessionKey(ctx, id, sessionKey)
}

func (s *orderService) ListClient(ctx context.Context, filter OrderFilter, sessionKey string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return s.orderRepository.ListBySessionKey(ctx, filter, sessionKey)
}

func (s *orderService) GetBranch(ctx context.Context, id string, branchID string) (*OrderDTO, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.BranchID != branchID {
		return nil, common.ErrOrderNotFound
	}
	return order, nil
}

func (s *orderService) ListBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return s.orderRepository.ListByBranch(ctx, filter, branchID)
}

func (s *orderService) UpdateBranch(ctx context.Context, id string, input OrderUpdateInput, branchID string) error {
	if input.OrderStatus == "" {
		return nil
	}
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if order.BranchID != branchID {
		return common.ErrOrderNotFound
	}
	return s.orderRepository.UpdateStatus(ctx, id, input.OrderStatus, input.CancellationReason)
}

func (s *orderService) GetAdmin(ctx context.Context, id string) (*OrderDTO, error) {
	return s.orderRepository.Get(ctx, id)
}

func (s *orderService) ListAdmin(ctx context.Context, filter OrderFilter) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return s.orderRepository.ListAll(ctx, filter)
}

func (s *orderService) UpdateAdmin(ctx context.Context, id string, input OrderUpdateInput) error {
	if input.OrderStatus == "" {
		return nil
	}
	return s.orderRepository.UpdateStatus(ctx, id, input.OrderStatus, input.CancellationReason)
}
