package folio

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"time"
)

type FolioService interface {
	// CreateForBooking opens a fresh folio for a booking (called at check-in).
	CreateForBooking(ctx context.Context, bookingID, branchID string) error
	// GetByBooking returns the folio for a booking (used by the room-order service).
	GetByBooking(ctx context.Context, bookingID string) (*BillDTO, error)
	// GetByBookingForBranch returns the folio, enforcing branch ownership (read).
	GetByBookingForBranch(ctx context.Context, bookingID, branchID string) (*BillDTO, error)
	// SetOrdersTotal updates the running total for an open folio.
	SetOrdersTotal(ctx context.Context, bookingID string, total float64) error
	// Settle marks an open folio settled with the chosen payment method.
	Settle(ctx context.Context, bookingID, branchID, role, settledBy, paymentMethod string) (*BillDTO, error)
	// HasOpenBalance reports whether the folio is open with a positive balance.
	HasOpenBalance(ctx context.Context, bookingID string) (bool, error)
}

type folioService struct {
	repository FolioRepository
	logger     config.Logger
}

func NewFolioService(repository FolioRepository, logger config.Logger) FolioService {
	return &folioService{repository: repository, logger: logger}
}

func (s *folioService) CreateForBooking(ctx context.Context, bookingID, branchID string) error {
	bill := &Bill{
		BookingID:     bookingID,
		BranchID:      branchID,
		Status:        string(BillOpen),
		Total:         0,
		PaymentMethod: "",
	}
	bill.ID = common.GenerateUUID()
	return s.repository.Create(ctx, bill)
}

func (s *folioService) GetByBooking(ctx context.Context, bookingID string) (*BillDTO, error) {
	bill, err := s.repository.GetByBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	dto := bill.ToDTO()
	return &dto, nil
}

func (s *folioService) GetByBookingForBranch(ctx context.Context, bookingID, branchID string) (*BillDTO, error) {
	bill, err := s.repository.GetByBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if bill.BranchID != branchID {
		return nil, common.ErrUnAuthorized
	}
	dto := bill.ToDTO()
	return &dto, nil
}

func (s *folioService) SetOrdersTotal(ctx context.Context, bookingID string, total float64) error {
	bill, err := s.repository.GetByBooking(ctx, bookingID)
	if err != nil {
		return err
	}
	// Never mutate a settled or voided folio.
	if bill.Status != string(BillOpen) {
		return nil
	}
	bill.Total = total
	return s.repository.Update(ctx, *bill)
}

func (s *folioService) Settle(ctx context.Context, bookingID, branchID, role, settledBy, paymentMethod string) (*BillDTO, error) {
	if !canSettleBills(role) {
		return nil, common.ErrUnAuthorized
	}
	bill, err := s.repository.GetByBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if bill.BranchID != branchID {
		return nil, common.ErrUnAuthorized
	}
	if bill.Status == string(BillSettled) {
		return nil, common.ErrBillAlreadySettled
	}
	bill.Status = string(BillSettled)
	bill.PaymentMethod = paymentMethod
	bill.SettledBy = sql.NullString{String: settledBy, Valid: settledBy != ""}
	bill.SettledAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := s.repository.Update(ctx, *bill); err != nil {
		return nil, err
	}
	dto := bill.ToDTO()
	return &dto, nil
}

func (s *folioService) HasOpenBalance(ctx context.Context, bookingID string) (bool, error) {
	bill, err := s.repository.GetByBooking(ctx, bookingID)
	if err != nil {
		// A booking without a folio (e.g. legacy) has no balance to settle.
		if err == common.ErrBillNotFound {
			return false, nil
		}
		return false, err
	}
	return bill.Status == string(BillOpen) && bill.Total > 0, nil
}
