package roomorder

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/menu"
	modoption "lazeez-core/internal/modifiers/option"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/folio"
	"lazeez-core/internal/roomsession"

	"golang.org/x/sync/errgroup"
)

type RoomOrderService interface {
	// Client: guest creates/reads room orders by room session key.
	CreateClient(ctx context.Context, order RoomOrderInput) (*CreateRoomOrderResponse, error)
	GetClient(ctx context.Context, id, sessionKey string) (*RoomOrderDTO, error)
	ListClient(ctx context.Context, filter RoomOrderFilter, sessionKey string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)

	// Branch: read for all branch roles; mutate only for front desk / room service.
	GetBranch(ctx context.Context, id, branchID string) (*RoomOrderDTO, error)
	ListBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
	ArchiveBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
	UpdateBranch(ctx context.Context, id string, input RoomOrderUpdateInput, role, branchID string) error

	// Admin: read all (super_admin) or merchant-scoped (super_branch_admin).
	GetAdmin(ctx context.Context, id string) (*RoomOrderDTO, error)
	ListAdmin(ctx context.Context, filter RoomOrderFilter) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
}

type roomOrderService struct {
	orderRepository       RoomOrderRepository
	menuService           menu.MenuService
	modifierOptionService modoption.ModifierOptionService
	roomSessionService    roomsession.RoomSessionService
	bookingRepository     booking.BookingRepository
	folioService          folio.FolioService
	logger                config.Logger
}

func NewRoomOrderService(
	orderRepository RoomOrderRepository,
	menuService menu.MenuService,
	modifierOptionService modoption.ModifierOptionService,
	roomSessionService roomsession.RoomSessionService,
	bookingRepository booking.BookingRepository,
	folioService folio.FolioService,
	logger config.Logger,
) RoomOrderService {
	return &roomOrderService{
		orderRepository:       orderRepository,
		menuService:           menuService,
		modifierOptionService: modifierOptionService,
		roomSessionService:    roomSessionService,
		bookingRepository:     bookingRepository,
		folioService:          folioService,
		logger:                logger,
	}
}

func (s *roomOrderService) CreateClient(ctx context.Context, order RoomOrderInput) (*CreateRoomOrderResponse, error) {
	itemsTotal := 0.0
	for _, oi := range order.OrderItems {
		itemsTotal += oi.Total
	}
	const epsilon = 0.01
	if order.Total+epsilon < itemsTotal {
		s.logger.Error("room order total is less than items total", "order_total", order.Total, "items_total", itemsTotal)
		return nil, common.ErrInvalidRequest
	}

	// Authenticate the guest via their room session (room QR + verified passcode).
	session, err := s.roomSessionService.GetValidForRoomOrder(ctx, order.SessionKey)
	if err != nil {
		return nil, err
	}

	// Re-confirm the booking is still active (guest may have been checked out).
	b, err := s.bookingRepository.Get(ctx, session.BookingID)
	if err != nil {
		return nil, err
	}
	if b.Status != string(booking.BookingActive) {
		return nil, common.ErrBookingNotActive
	}

	bill, err := s.folioService.GetByBooking(ctx, session.BookingID)
	if err != nil {
		return nil, err
	}
	if bill.Status != folio.BillOpen {
		return nil, common.ErrBillUnsettled
	}

	if err := s.validateOrderBeforeCreate(ctx, session.BranchID, order.OrderItems); err != nil {
		return nil, err
	}

	orderNumber, err := s.assignOrderNumber(ctx, session.BranchID)
	if err != nil {
		s.logger.Error("failed to assign room order number", "error", err)
		return nil, common.ErrInternalServerError
	}

	orderModel := RoomOrder{
		OrderNumber: orderNumber,
		RoomID:      session.RoomID,
		BookingID:   session.BookingID,
		BranchID:    session.BranchID,
		SessionKey:  session.SessionKey,
		OrderStatus: string(StatusPending),
		Total:       order.Total,
		BillID:      bill.ID,
	}
	orderModel.ID = common.GenerateUUID()

	created, err := s.orderRepository.Create(ctx, orderModel)
	if err != nil {
		return nil, err
	}

	g, itemCtx := errgroup.WithContext(ctx)
	for _, oi := range order.OrderItems {
		oi := oi
		g.Go(func() error {
			return s.orderRepository.CreateItem(itemCtx, RoomOrderItem{
				Base:            common.Base{ID: common.GenerateUUID()},
				RoomOrderID:     created.ID,
				MenuItemID:      oi.MenuItemID,
				ModifierOptions: oi.ModifierOptions,
				Quantity:        oi.Quantity,
				Price:           oi.Price,
				Total:           oi.Total,
			})
		})
	}
	if err := g.Wait(); err != nil {
		s.logger.Error("failed to create room order items", "error", err)
		if delErr := s.orderRepository.Delete(ctx, created.ID); delErr != nil {
			s.logger.Error("failed to roll back room order after item failure", "order_id", created.ID, "error", delErr)
		}
		return nil, err
	}

	// Add the new order to the running room bill total.
	if err := s.recomputeFolio(ctx, session.BookingID); err != nil {
		s.logger.Error("failed to update folio total after room order create", "booking_id", session.BookingID, "error", err)
	}

	return &CreateRoomOrderResponse{
		ID:          created.ID,
		OrderNumber: created.OrderNumber,
		Total:       created.Total,
		OrderStatus: created.OrderStatus,
		BillID:      created.BillID,
	}, nil
}

// recomputeFolio recalculates the booking's room bill total from non-cancelled orders.
func (s *roomOrderService) recomputeFolio(ctx context.Context, bookingID string) error {
	total, err := s.orderRepository.SumActiveByBooking(ctx, bookingID)
	if err != nil {
		return err
	}
	return s.folioService.SetOrdersTotal(ctx, bookingID, total)
}

func (s *roomOrderService) GetClient(ctx context.Context, id, sessionKey string) (*RoomOrderDTO, error) {
	return s.orderRepository.GetBySessionKey(ctx, id, sessionKey)
}

func (s *roomOrderService) ListClient(ctx context.Context, filter RoomOrderFilter, sessionKey string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	return s.orderRepository.ListBySessionKey(ctx, filter, sessionKey)
}

func (s *roomOrderService) GetBranch(ctx context.Context, id, branchID string) (*RoomOrderDTO, error) {
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.BranchID != branchID {
		return nil, common.ErrOrderNotFound
	}
	return order, nil
}

func (s *roomOrderService) ListBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	return s.orderRepository.ListByBranch(ctx, filter, branchID)
}

func (s *roomOrderService) ArchiveBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	if err := ValidateArchiveDateRange(filter); err != nil {
		return nil, err
	}
	if filter.DateFrom == nil || filter.DateTo == nil {
		return nil, common.ErrInvalidRequest
	}
	return s.orderRepository.ListArchiveByBranch(ctx, filter, branchID)
}

func (s *roomOrderService) UpdateBranch(ctx context.Context, id string, input RoomOrderUpdateInput, role, branchID string) error {
	if !canMutateRoomOrders(role) {
		return common.ErrUnAuthorized
	}
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if order.BranchID != branchID {
		return common.ErrOrderNotFound
	}
	return s.updateOrderStatus(ctx, order, input)
}

func (s *roomOrderService) GetAdmin(ctx context.Context, id string) (*RoomOrderDTO, error) {
	return s.orderRepository.Get(ctx, id)
}

func (s *roomOrderService) ListAdmin(ctx context.Context, filter RoomOrderFilter) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	return s.orderRepository.ListAll(ctx, filter)
}
