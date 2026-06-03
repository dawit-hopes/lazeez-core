package booking

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/rooms/folio"
	"lazeez-core/internal/rooms/room"

	"golang.org/x/sync/errgroup"
)

const passcodeLength = 6

type BookingService interface {
	CreateBooking(ctx context.Context, req BookingRequestDTO, role, branchID, merchantID string) (*BookingDTO, error)
	GetBooking(ctx context.Context, id, role, branchID, merchantID string) (*BookingDTO, error)
	ListBookings(ctx context.Context, filter common.Filter, role, branchID, merchantID string) (*common.PaginatedResponse[[]*BookingDTO], error)
	UpdateBooking(ctx context.Context, id string, req BookingUpdateRequestDTO, role, branchID, merchantID string) (*BookingDTO, error)
	CheckOut(ctx context.Context, id, role, branchID, merchantID string) (*BookingDTO, error)
	Cancel(ctx context.Context, id, role, branchID, merchantID string) (*BookingDTO, error)
	DeleteBooking(ctx context.Context, id, role, branchID, merchantID string) error
	ReGeneratePassCode(ctx context.Context, id, role, branchID, merchantID string) (*string, error)
	GetBookingById(ctx context.Context, id string) (*BookingDTO, error)
	GetBookingByRoom(ctx context.Context, roomID string) (*Booking, error)
	VerifyGuestPasscode(ctx context.Context, b *Booking, passcode string) error
}

type bookingService struct {
	repository      BookingRepository
	roomRepo        room.RoomRepository
	branchService   branch.BranchService
	merchantService merchant.MerchantService
	keyService      key.KeyService
	folioService    folio.FolioService
	logger          config.Logger
}

func NewBookingService(
	repository BookingRepository,
	roomRepo room.RoomRepository,
	branchService branch.BranchService,
	merchantService merchant.MerchantService,
	keyService key.KeyService,
	folioService folio.FolioService,
	logger config.Logger,
) BookingService {
	return &bookingService{
		repository:      repository,
		roomRepo:        roomRepo,
		branchService:   branchService,
		merchantService: merchantService,
		keyService:      keyService,
		folioService:    folioService,
		logger:          logger,
	}
}

// authorizeManage enforces who may mutate a specific booking: only front desk agents,
// and only within their own branch.
func (s *bookingService) authorizeManage(b *Booking, role, branchID string) error {
	if !canMutateBookings(role) {
		return common.ErrUnAuthorized
	}
	if branchID != "" && b.BranchID == branchID {
		return nil
	}
	return common.ErrUnAuthorized
}

func (s *bookingService) CreateBooking(ctx context.Context, req BookingRequestDTO, role, branchID, merchantID string) (*BookingDTO, error) {
	if !canMutateBookings(role) {
		return nil, common.ErrUnAuthorized
	}
	if branchID == "" {
		return nil, common.ErrUnAuthorized
	}

	g, ctxg := errgroup.WithContext(ctx)
	rm := &room.RoomWithType{}
	var (
		plainPasscode  string
		hashedPasscode string
	)

	g.Go(func() error {
		var err error
		rm, err = s.roomRepo.Get(ctxg, req.RoomID)
		if err != nil {
			return err
		}
		*rm = *rm
		return nil
	})

	g.Go(func() error {
		occupied, err := s.repository.HasActiveBooking(ctxg, req.RoomID)
		if err != nil {
			return err
		}
		if occupied {
			return common.ErrRoomOccupied
		}
		return nil
	})

	g.Go(func() error {
		var err error
		plainPasscode, hashedPasscode, err = s.generatePasscode()
		if err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		s.logger.Error("failed to create booking", "error", err)
		return nil, err
	}

	b := req.ToModel()
	b.ID = common.GenerateUUID()
	b.BranchID = rm.BranchID
	b.Status = string(BookingActive)
	b.PasscodeHash = hashedPasscode
	b.CheckOutDate = b.CheckInDate.AddDate(0, 0, b.NumberOfNights)


	// ====================== make this atomic ======================
	if err := s.repository.Create(ctx, &b); err != nil {
		return nil, err
	}

	// Open a fresh folio for the stay so room orders can be charged to it.
	if err := s.folioService.CreateForBooking(ctx, b.ID, b.BranchID); err != nil {
		s.logger.Error("failed to open folio for booking", "booking_id", b.ID, "error", err)
		if delErr := s.repository.Delete(ctx, b.ID); delErr != nil {
			s.logger.Error("failed to roll back booking after folio failure", "booking_id", b.ID, "error", delErr)
		}
		return nil, common.ErrInternalServerError
	}

	// make the room status occupied
	if err := s.roomRepo.UpdateStatus(ctx, b.RoomID, string(room.RoomStatusOccupied)); err != nil {
		s.logger.Error("failed to update room status", "room_id", b.RoomID, "error", err)
		if delErr := s.repository.Delete(ctx, b.ID); delErr != nil {
			s.logger.Error("failed to roll back booking after room status update failure", "booking_id", b.ID, "error", delErr)
		}
		return nil, common.ErrInternalServerError
	}

	// ====================== make this atomic ======================
	dto := b.ToDTO()
	dto.Passcode = plainPasscode
	return &dto, nil
}

func (s *bookingService) GetBooking(ctx context.Context, id, role, branchID, merchantID string) (*BookingDTO, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeRead(ctx, b, role, branchID, merchantID); err != nil {
		return nil, err
	}
	dto := b.ToDTO()
	return &dto, nil
}

func (s *bookingService) ListBookings(ctx context.Context, filter common.Filter, role, branchID, merchantID string) (*common.PaginatedResponse[[]*BookingDTO], error) {
	listBranchID, listMerchantID, err := s.resolveListScope(role, branchID, merchantID, filter)
	if err != nil {
		return nil, err
	}

	result, err := s.repository.List(ctx, filter, listBranchID, listMerchantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*BookingDTO, len(result.Data))
	for i, b := range result.Data {
		dto := b.ToDTO()
		dtos[i] = &dto
	}
	return &common.PaginatedResponse[[]*BookingDTO]{
		Data: dtos,
		Meta: result.Meta,
	}, nil
}

func (s *bookingService) UpdateBooking(ctx context.Context, id string, req BookingUpdateRequestDTO, role, branchID, merchantID string) (*BookingDTO, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeManage(b, role, branchID); err != nil {
		return nil, err
	}

	if req.GuestName != "" {
		b.GuestName = req.GuestName
	}
	if req.GuestPhone != "" {
		b.GuestPhone = req.GuestPhone
	}

	if err := s.repository.Update(ctx, *b); err != nil {
		return nil, err
	}
	dto := b.ToDTO()
	return &dto, nil
}

func (s *bookingService) CheckOut(ctx context.Context, id, role, branchID, merchantID string) (*BookingDTO, error) {
	return s.transition(ctx, id, role, branchID, BookingCheckedOut)
}

func (s *bookingService) Cancel(ctx context.Context, id, role, branchID, merchantID string) (*BookingDTO, error) {
	return s.transition(ctx, id, role, branchID, BookingCancelled)
}

// transition moves an active booking to a terminal status (checked_out or cancelled).
func (s *bookingService) transition(ctx context.Context, id, role, branchID string, target BookingStatus) (*BookingDTO, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeManage(b, role, branchID); err != nil {
		return nil, err
	}
	if b.Status != string(BookingActive) {
		return nil, common.ErrBookingNotActive
	}
	// A guest cannot be checked out while the room bill still has an open balance.
	if target == BookingCheckedOut {
		hasBalance, err := s.folioService.HasOpenBalance(ctx, b.ID)
		if err != nil {
			return nil, err
		}
		if hasBalance {
			return nil, common.ErrBillUnsettled
		}
	}
	b.Status = string(target)
	if err := s.repository.Update(ctx, *b); err != nil {
		return nil, err
	}
	if target == BookingCheckedOut || target == BookingCancelled {
		if err := s.roomRepo.UpdateStatus(ctx, b.RoomID, string(room.RoomStatusVacant)); err != nil {
			s.logger.Error("failed to update room status", "room_id", b.RoomID, "error", err)
			return nil, err
		}
	}
	dto := b.ToDTO()
	return &dto, nil
}

func (s *bookingService) DeleteBooking(ctx context.Context, id, role, branchID, merchantID string) error {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeManage(b, role, branchID); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func (s *bookingService) ReGeneratePassCode(ctx context.Context, id, role, branchID, merchantID string) (*string, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		s.logger.Error("failed to get booking", "error", err)
		return nil, err
	}
	if err := s.authorizeManage(b, role, branchID); err != nil {
		s.logger.Error("failed to authorize manage", "error", err)
		return nil, err
	}
	plainPasscode, hashedPasscode, err := s.generatePasscode()
	if err != nil {
		return nil, err
	}
	if err := s.repository.SetPasscodeAttempts(ctx, b.ID, 0, sql.NullTime{}); err != nil {
		s.logger.Error("failed to reset passcode security on regenerate", "booking_id", b.ID, "error", err)
		return nil, err
	}
	if err := s.repository.UpdatePasscode(ctx, b.ID, hashedPasscode); err != nil {
		s.logger.Error("failed to update passcode hash on regenerate", "booking_id", b.ID, "error", err)
		return nil, err
	}
	return &plainPasscode, nil
}

func (s *bookingService) generatePasscode() (string, string, error) {
	plainPasscode := common.GeneratePasscode(passcodeLength)
	hashedPasscode, err := s.keyService.HashPassword(plainPasscode)
	if err != nil {
		s.logger.Error("failed to hash passcode", "error", err)
		return "", "", common.ErrInternalServerError
	}
	return plainPasscode, hashedPasscode, nil
}

func (s *bookingService) GetBookingById(ctx context.Context, id string) (*BookingDTO, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := b.ToDTO()
	return &dto, nil
}

func (s *bookingService) GetBookingByRoom(ctx context.Context, roomID string) (*Booking, error) {
	b, err := s.repository.GetActiveByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return b, nil
}