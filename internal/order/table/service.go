package order

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/clientsession"
	"lazeez-core/internal/common"
	item "lazeez-core/internal/order/Item"
	"lazeez-core/internal/menu/dinning"
	"lazeez-core/internal/payment"
	option "lazeez-core/internal/modifiers/option"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	SUCCESSFUL_PAYMENT_STATUS = "success"
	FAILED_PAYMENT_STATUS     = "failed"
	ETB                       = "ETB"
	EMAIL                     = "dawit.abrahame@gmail.com"
)

type OrderService interface {
	// Client: create order (no auth), get/list by sessionKey
	CreateClient(ctx context.Context, order OrderInput) (*CreateOrderResponse, error)
	GetClient(ctx context.Context, id string, sessionKey string) (*OrderDTO, error)
	ListClient(ctx context.Context, filter OrderFilter, sessionKey string) (*common.PaginatedResponse[[]*OrderDTO], error)
	ProcessPaymentWebHook(ctx context.Context, payload PaymentWebHookPayload, rawBody []byte, chapaSignature, xChapaSignature string) error
	CancelPaymentClient(ctx context.Context, orderID string, sessionKey string) error

	// Branch: get/list/update orders for branch (branch_id from token)
	GetBranch(ctx context.Context, id string, branchID string) (*OrderDTO, error)
	ListBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error)
	ArchiveBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error)
	UpdateBranch(ctx context.Context, id string, input OrderUpdateInput, branchID string) error

	// Admin: get/list/update all orders
	GetAdmin(ctx context.Context, id string) (*OrderDTO, error)
	ListAdmin(ctx context.Context, filter OrderFilter) (*common.PaginatedResponse[[]*OrderDTO], error)
	UpdateAdmin(ctx context.Context, id string, input OrderUpdateInput) error
}

type orderService struct {
	orderRepository      OrderRepository
	orderItemService     item.OrderItemService
	menuService          menu.MenuService
	modifierOptionService option.ModifierOptionService
	clientSessionService clientsession.ClientSessionService
	paymentService       payment.PaymentService
	logger               config.Logger
	callbackURL          string
	menuBaseURL          string
}

func NewOrderService(
	orderRepository OrderRepository,
	orderItemService item.OrderItemService,
	menuService menu.MenuService,
	modifierOptionService option.ModifierOptionService,
	clientSessionService clientsession.ClientSessionService,
	paymentService payment.PaymentService,
	logger config.Logger,
	callbackURL string,
	menuBaseURL string,
) OrderService {
	return &orderService{
		orderRepository:       orderRepository,
		orderItemService:      orderItemService,
		menuService:           menuService,
		modifierOptionService: modifierOptionService,
		clientSessionService:  clientSessionService,
		paymentService:        paymentService,
		logger:                logger,
		callbackURL:           callbackURL,
		menuBaseURL:           menuBaseURL,
	}
}

func (s *orderService) CreateClient(ctx context.Context, order OrderInput) (*CreateOrderResponse, error) {
	itemsTotal := 0.0
	for _, oi := range order.OrderItems {
		itemsTotal += oi.Total
	}
	const epsilon = 0.01
	if order.Total+epsilon < itemsTotal {
		s.logger.Error("order total is less than items total", "order total", order.Total, "items total", itemsTotal)
		return nil, common.ErrInvalidRequest
	}

	branchID, tableReference, err := s.resolveOrderContext(ctx, order.BranchID, order.SessionKey)
	if err != nil {
		s.logger.Error("failed to resolve order context", "error", err)
		return nil, common.ErrInvalidRequest
	}

	if err := s.validateOrderBeforeCreate(ctx, branchID, order.OrderItems); err != nil {
		return nil, err
	}

	orderNumber, err := s.assignOrderNumber(ctx, branchID)
	if err != nil {
		s.logger.Error("failed to assign order number", "error", err)
		return nil, common.ErrInternalServerError
	}

	paymentMethod := order.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "mobile"
	}

	orderModel := Order{
		OrderNumber:     orderNumber,
		TableNumber:     order.TableNumber,
		BranchID:        branchID,
		SessionKey:      order.SessionKey,
		Total:           order.Total,
		OrderStatus:     string(StatusPending),
		PaymentStatus:   PaymentStatusPending,
		PaymentMethod:   paymentMethod,
		PaymentDate:     time.Now(),
		PaymentAmount:   order.Total,
		PaymentCurrency: ETB,
	}
	orderModel.ID = common.GenerateUUID()
	created, err := s.orderRepository.Create(ctx, orderModel)
	if err != nil {
		s.logger.Error("failed to create order", "error", err)
		return nil, common.ErrInternalServerError
	}

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

	paymentPayload := payment.PaymentPayload{
		Amount:      strconv.FormatFloat(order.Total, 'f', -1, 64),
		Currency:    ETB,
		Email:       EMAIL,
		FirstName:   "Guest",
		LastName:    "Customer",
		PhoneNumber: "0900000000",
		TxRef:       created.ID,
		CallbackURL: s.callbackURL,
		ReturnURL:   common.BuildPaymentReturnURL(s.menuBaseURL, tableReference, created.ID),
		Customization: map[string]any{
			"title":       "Order Payment",
			"description": "Payment for order " + strconv.Itoa(created.OrderNumber),
		},
		Meta: map[string]any{
			"order_id": created.ID,
		},
	}

	initializationResponse, err := s.paymentService.InitializePayment(ctx, paymentPayload)
	if err != nil {
		s.logger.Error("failed to initialize payment", "error", err)
		if delErr := s.orderRepository.Delete(ctx, created.ID); delErr != nil {
			s.logger.Error("failed to rollback order after payment init failure", "order_id", created.ID, "error", delErr)
		}
		return nil, common.ErrInternalServerError
	}

	return &CreateOrderResponse{
		ID:            created.ID,
		OrderNumber:   created.OrderNumber,
		Total:         created.Total,
		OrderStatus:   created.OrderStatus,
		PaymentStatus: created.PaymentStatus,
		CheckoutURL:   initializationResponse.CheckoutURL,
	}, nil
}

func (s *orderService) ProcessPaymentWebHook(ctx context.Context, payload PaymentWebHookPayload, rawBody []byte, chapaSignature, xChapaSignature string) error {
	if !s.paymentService.VerifySignature(rawBody, chapaSignature, xChapaSignature) {
		s.logger.Error("failed to verify webhook signature")
		return common.ErrUnAuthorized
	}

	txRef := payload.TransactionRef()
	if txRef == "" {
		s.logger.Error("webhook missing transaction reference")
		return common.ErrInvalidRequest
	}

	existingOrder, err := s.orderRepository.Get(ctx, txRef)
	if err != nil {
		return err
	}

	switch payload.Status {
	case SUCCESSFUL_PAYMENT_STATUS:
		if existingOrder.PaymentStatus == PaymentStatusSuccess {
			return nil
		}
	case FAILED_PAYMENT_STATUS:
		if existingOrder.PaymentStatus == PaymentStatusFailed {
			return nil
		}
	default:
		s.logger.Error("unknown payment status", "status", payload.Status)
		return common.ErrInvalidRequest
	}

	verification, err := s.paymentService.VerifyTransaction(ctx, txRef)
	if err != nil {
		s.logger.Error("failed to verify transaction", "error", err)
		return common.ErrInternalServerError
	}

	if verification.TxRef != "" && verification.TxRef != txRef {
		s.logger.Error("transaction reference mismatch", "expected", txRef, "actual", verification.TxRef)
		return common.ErrInvalidRequest
	}

	if err := s.validateVerifiedAmount(existingOrder.Total, verification.Amount, verification.Currency); err != nil {
		return err
	}

	chapaReference := payload.ChapaReference()
	verifiedAmount, _ := strconv.ParseFloat(verification.Amount, 64)

	switch payload.Status {
	case SUCCESSFUL_PAYMENT_STATUS:
		return s.orderRepository.UpdatePaymentStatus(ctx, txRef, PaymentStatusUpdate{
			PaymentStatus:        PaymentStatusSuccess,
			PaymentTransactionID: chapaReference,
			PaymentAmount:        verifiedAmount,
			PaymentCurrency:      verification.Currency,
		})
	case FAILED_PAYMENT_STATUS:
		return s.orderRepository.UpdatePaymentStatus(ctx, txRef, PaymentStatusUpdate{
			PaymentStatus:        PaymentStatusFailed,
			PaymentTransactionID: chapaReference,
			PaymentAmount:        verifiedAmount,
			PaymentCurrency:      verification.Currency,
			OrderStatus:          string(StatusCancelled),
			CancellationReason:   "Payment failed",
		})
	}

	return common.ErrInternalServerError
}

func (s *orderService) validateVerifiedAmount(expectedTotal float64, verifiedAmount, verifiedCurrency string) error {
	if verifiedCurrency != "" && verifiedCurrency != ETB {
		s.logger.Error("unexpected payment currency", "currency", verifiedCurrency)
		return common.ErrInvalidRequest
	}

	amount, err := strconv.ParseFloat(verifiedAmount, 64)
	if err != nil {
		s.logger.Error("failed to parse verified amount", "amount", verifiedAmount, "error", err)
		return common.ErrInvalidRequest
	}

	const epsilon = 0.01
	if amount < expectedTotal-epsilon || amount > expectedTotal+epsilon {
		s.logger.Error("verified payment amount mismatch", "expected", expectedTotal, "actual", amount)
		return common.ErrInvalidRequest
	}

	return nil
}

func (s *orderService) CancelPaymentClient(ctx context.Context, orderID, sessionKey string) error {
	order, err := s.orderRepository.GetBySessionKey(ctx, orderID, sessionKey)
	if err != nil {
		return err
	}

	if order.PaymentStatus == PaymentStatusSuccess {
		return common.ErrInvalidRequest
	}

	if order.PaymentStatus == PaymentStatusFailed {
		return nil
	}

	if cancelErr := s.paymentService.CancelTransaction(ctx, orderID); cancelErr != nil {
		s.logger.Warn("failed to cancel chapa transaction", "order_id", orderID, "error", cancelErr)
	}

	return s.orderRepository.UpdatePaymentStatus(ctx, orderID, PaymentStatusUpdate{
		PaymentStatus:        PaymentStatusFailed,
		OrderStatus:          string(StatusCancelled),
		CancellationReason:   "Payment cancelled",
	})
}

func (s *orderService) resolveOrderContext(ctx context.Context, branchID, sessionKey string) (string, string, error) {
	if sessionKey == "" {
		s.logger.Error("session key is required for client orders")
		return "", "", common.ErrInvalidRequest
	}

	session, err := s.clientSessionService.GetValidForOrder(ctx, sessionKey)
	if err != nil {
		return "", "", err
	}

	if branchID != "" && branchID != session.BranchID {
		s.logger.Error("branch id does not match client session", "branch_id", branchID, "session_branch_id", session.BranchID)
		return "", "", common.ErrInvalidRequest
	}

	return session.BranchID, session.TableReference, nil
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

func (s *orderService) ArchiveBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	if err := ValidateArchiveDateRange(filter); err != nil {
		return nil, err
	}
	if filter.DateFrom == nil || filter.DateTo == nil {
		return nil, common.ErrInvalidRequest
	}
	return s.orderRepository.ListArchiveByBranch(ctx, filter, branchID)
}

func (s *orderService) UpdateBranch(ctx context.Context, id string, input OrderUpdateInput, branchID string) error {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if order.BranchID != branchID {
		return common.ErrOrderNotFound
	}
	return s.updateOrderStatus(ctx, id, input)
}

func (s *orderService) GetAdmin(ctx context.Context, id string) (*OrderDTO, error) {
	return s.orderRepository.Get(ctx, id)
}

func (s *orderService) ListAdmin(ctx context.Context, filter OrderFilter) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return s.orderRepository.ListAll(ctx, filter)
}

func (s *orderService) UpdateAdmin(ctx context.Context, id string, input OrderUpdateInput) error {
	return s.updateOrderStatus(ctx, id, input)
}
