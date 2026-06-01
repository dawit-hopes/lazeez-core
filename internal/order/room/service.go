package roomorder

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	menu "lazeez-core/internal/menu/dinning"
	modoption "lazeez-core/internal/modifiers/option"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/folio"
	"lazeez-core/internal/rooms/room"

	"golang.org/x/sync/errgroup"
)

type RoomOrderService interface {
	// Client: guest creates/reads room orders by room reference and pass code.
	CreateClient(ctx context.Context, order RoomOrderInput) (*CreateRoomOrderResponse, error)
	GetClient(ctx context.Context, id, reference, passCode string) (*RoomOrderDTO, error)
	ListClient(ctx context.Context, filter RoomOrderFilter, reference, passCode string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)

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
	folioService          folio.FolioService
	roomService           room.RoomService
	bookingService        booking.BookingService
	logger                config.Logger
}

func NewRoomOrderService(
	orderRepository RoomOrderRepository,
	menuService menu.MenuService,
	modifierOptionService modoption.ModifierOptionService,
	folioService folio.FolioService,
	roomService room.RoomService,
	bookingService booking.BookingService,
	logger config.Logger,
) RoomOrderService {
	return &roomOrderService{
		orderRepository:       orderRepository,
		menuService:           menuService,
		modifierOptionService: modifierOptionService,
		folioService:          folioService,
		roomService:           roomService,
		bookingService:        bookingService,
		logger:                logger,
	}
}

func (s *roomOrderService) validateBalance(order *RoomOrderInput) error {
	itemsTotal := 0.0
	for _, oi := range order.OrderItems {
		itemsTotal += oi.Total
	}
	const epsilon = 0.01
	if order.Total+epsilon < itemsTotal {
		s.logger.Error("room order total is less than items total", "order_total", order.Total, "items_total", itemsTotal)
		return common.ErrInvalidRequest
	}
	return nil
}

func (s *roomOrderService) validateRoom(ctx context.Context, roomID string, branchID string) error {
	rm, err := s.roomService.GetRoomById(ctx, roomID)
	if err != nil {
		return err
	}

	if rm.BranchID != branchID {
		return common.ErrUnAuthorized
	}

	if rm.Status != room.RoomStatusOccupied {
		return common.ErrRoomNotOccupied
	}

	return nil
}

func (s *roomOrderService) validateBooking(ctx context.Context, roomID string) (*booking.Booking, error) {
	b, err := s.bookingService.GetBookingByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if b.Status != string(booking.BookingActive) {
		return nil, common.ErrBookingNotActive
	}

	return b, nil
}

func (s *roomOrderService) resolveGuestBooking(ctx context.Context, reference, passCode string) (*booking.Booking, error) {
	rm, err := s.roomService.GetRoomByReference(ctx, reference)
	if err != nil {
		return nil, err
	}

	if err := s.validateRoom(ctx, rm.ID, rm.BranchID); err != nil {
		return nil, err
	}

	activeBooking, err := s.validateBooking(ctx, rm.ID)
	if err != nil {
		return nil, err
	}

	if err := s.bookingService.VerifyGuestPasscode(ctx, activeBooking, passCode); err != nil {
		return nil, err
	}

	return activeBooking, nil
}

func (s *roomOrderService) requireOpenFolio(ctx context.Context, bookingID string) (*folio.BillDTO, error) {
	bill, err := s.folioService.GetByBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if bill.Status != folio.BillOpen {
		return nil, common.ErrBillUnsettled
	}
	return bill, nil
}

// recomputeFolio recalculates the booking's room bill total from non-cancelled orders.
func (s *roomOrderService) recomputeFolio(ctx context.Context, bookingID string) error {
	total, err := s.orderRepository.SumActiveByBooking(ctx, bookingID)
	if err != nil {
		return err
	}
	return s.folioService.SetOrdersTotal(ctx, bookingID, total)
}

func (s *roomOrderService) CreateClient(ctx context.Context, order RoomOrderInput) (*CreateRoomOrderResponse, error) {
	if err := s.validateBalance(&order); err != nil {
		return nil, err
	}

	rm, err := s.roomService.GetRoomByReference(ctx, order.Reference)
	if err != nil {
		return nil, err
	}

	roomID := rm.ID
	branchID := rm.BranchID

	if err := s.validateRoom(ctx, roomID, branchID); err != nil {
		return nil, err
	}

	activeBooking, err := s.validateBooking(ctx, roomID)
	if err != nil {
		return nil, err
	}

	bookingID := activeBooking.ID

	if err := s.bookingService.VerifyGuestPasscode(ctx, activeBooking, order.PassCode); err != nil {
		return nil, err
	}

	bill, err := s.requireOpenFolio(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	orderNumber, err := s.assignOrderNumber(ctx, branchID)
	if err != nil {
		s.logger.Error("failed to assign room order number", "error", err)
		return nil, common.ErrInternalServerError
	}

	if err := s.validateOrderBeforeCreate(ctx, branchID, order.OrderItems); err != nil {
		return nil, err
	}

	orderModel := RoomOrder{
		OrderNumber: orderNumber,
		RoomID:      roomID,
		BookingID:   bookingID,
		BranchID:    branchID,
		BillID:      bill.ID,
		OrderStatus: string(StatusPending),
		Total:       order.Total,
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

	if err := s.recomputeFolio(ctx, bookingID); err != nil {
		s.logger.Error("failed to update folio total after room order create", "booking_id", bookingID, "error", err)
		if delErr := s.orderRepository.Delete(ctx, created.ID); delErr != nil {
			s.logger.Error("failed to roll back room order after folio failure", "order_id", created.ID, "error", delErr)
		}
		return nil, common.ErrInternalServerError
	}

	return &CreateRoomOrderResponse{
		ID:          created.ID,
		OrderNumber: created.OrderNumber,
		Total:       created.Total,
		OrderStatus: created.OrderStatus,
		BillID:      created.BillID,
	}, nil
}

func (s *roomOrderService) GetClient(ctx context.Context, id, reference, passCode string) (*RoomOrderDTO, error) {
	activeBooking, err := s.resolveGuestBooking(ctx, reference, passCode)
	if err != nil {
		return nil, err
	}
	return s.orderRepository.GetByBookingID(ctx, id, activeBooking.ID)
}

func (s *roomOrderService) ListClient(ctx context.Context, filter RoomOrderFilter, reference, passCode string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	activeBooking, err := s.resolveGuestBooking(ctx, reference, passCode)
	if err != nil {
		return nil, err
	}
	return s.orderRepository.ListByBookingID(ctx, filter, activeBooking.ID)
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
