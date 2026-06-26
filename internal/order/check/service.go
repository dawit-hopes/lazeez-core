package check

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"time"
)

type CheckService interface {
	// GetOrCreateOpenCheck returns the id of the open check for (branch, table), creating one if needed.
	GetOrCreateOpenCheck(ctx context.Context, branchID string, tableNumber int) (string, error)
	// Recompute refreshes an open check's bill math (subtotal + VAT + service charge + total).
	Recompute(ctx context.Context, checkID string) error
	// GetForBranch returns a check with its derived readiness, enforcing branch ownership.
	GetForBranch(ctx context.Context, checkID, branchID string) (*CheckReadinessDTO, error)
	// ListOpenForBranch lists open checks (with readiness) for the cashier.
	ListOpenForBranch(ctx context.Context, branchID string) ([]*CheckReadinessDTO, error)
	// Settle closes a check (cash-only) once the system says it is ready for billing.
	Settle(ctx context.Context, checkID, branchID, role, settledBy, paymentMethod string) (*CheckDTO, error)
}

type checkService struct {
	repository CheckRepository
	logger     config.Logger
}

func NewCheckService(repository CheckRepository, logger config.Logger) CheckService {
	return &checkService{repository: repository, logger: logger}
}

func (s *checkService) GetOrCreateOpenCheck(ctx context.Context, branchID string, tableNumber int) (string, error) {
	existing, err := s.repository.GetOpenByTable(ctx, branchID, tableNumber)
	if err == nil {
		return existing.ID, nil
	}
	if err != common.ErrCheckNotFound {
		return "", err
	}

	c := &TableCheck{
		BranchID:    branchID,
		TableNumber: tableNumber,
		Status:      string(CheckOpen),
	}
	c.ID = common.GenerateUUID()
	if createErr := s.repository.Create(ctx, c); createErr != nil {
		// A concurrent waiter may have opened the check first (partial unique index);
		// fall back to reading the existing open check.
		if again, getErr := s.repository.GetOpenByTable(ctx, branchID, tableNumber); getErr == nil {
			return again.ID, nil
		}
		return "", createErr
	}
	return c.ID, nil
}

func (s *checkService) Recompute(ctx context.Context, checkID string) error {
	c, err := s.repository.GetByID(ctx, checkID)
	if err != nil {
		return err
	}
	// Never mutate a closed or voided check.
	if c.Status != string(CheckOpen) {
		return nil
	}

	subtotal, vatPercent, serviceChargePercent, err := s.repository.ComputeBill(ctx, checkID, c.BranchID)
	if err != nil {
		return err
	}
	c.Subtotal = subtotal
	c.VatAmount = subtotal * vatPercent / 100
	c.ServiceChargeAmount = subtotal * serviceChargePercent / 100
	c.Total = c.Subtotal + c.VatAmount + c.ServiceChargeAmount
	return s.repository.Update(ctx, *c)
}

func (s *checkService) GetForBranch(ctx context.Context, checkID, branchID string) (*CheckReadinessDTO, error) {
	c, err := s.repository.GetByID(ctx, checkID)
	if err != nil {
		return nil, err
	}
	if c.BranchID != branchID {
		return nil, common.ErrCheckNotFound
	}
	return s.withReadiness(ctx, c)
}

func (s *checkService) ListOpenForBranch(ctx context.Context, branchID string) ([]*CheckReadinessDTO, error) {
	checks, err := s.repository.ListOpenByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}
	policy, err := s.repository.GetBillPrintPolicy(ctx, branchID)
	if err != nil {
		return nil, err
	}
	out := make([]*CheckReadinessDTO, 0, len(checks))
	for _, c := range checks {
		counts, err := s.repository.GetReadiness(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		dto := &CheckReadinessDTO{
			CheckDTO:  c.ToDTO(),
			Readiness: deriveReadiness(counts, policy),
			Counts:    counts,
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *checkService) Settle(ctx context.Context, checkID, branchID, role, settledBy, paymentMethod string) (*CheckDTO, error) {
	if !canSettleChecks(role) {
		return nil, common.ErrUnAuthorized
	}
	c, err := s.repository.GetByID(ctx, checkID)
	if err != nil {
		return nil, err
	}
	if c.BranchID != branchID {
		return nil, common.ErrCheckNotFound
	}
	if c.Status != string(CheckOpen) {
		return nil, common.ErrCheckNotOpen
	}

	// Refresh totals before gating, so the cashier settles the current bill.
	if err := s.Recompute(ctx, checkID); err != nil {
		return nil, err
	}
	c, err = s.repository.GetByID(ctx, checkID)
	if err != nil {
		return nil, err
	}

	policy, err := s.repository.GetBillPrintPolicy(ctx, branchID)
	if err != nil {
		return nil, err
	}
	counts, err := s.repository.GetReadiness(ctx, checkID)
	if err != nil {
		return nil, err
	}
	if deriveReadiness(counts, policy) != ReadinessReadyForBilling {
		return nil, common.ErrCheckNotReady
	}

	c.Status = string(CheckClosed)
	c.PaymentMethod = paymentMethod
	c.SettledBy = sql.NullString{String: settledBy, Valid: settledBy != ""}
	c.SettledAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := s.repository.Update(ctx, *c); err != nil {
		return nil, err
	}
	if err := s.repository.MarkOrdersServed(ctx, checkID); err != nil {
		return nil, err
	}

	dto := c.ToDTO()
	return &dto, nil
}

func (s *checkService) withReadiness(ctx context.Context, c *TableCheck) (*CheckReadinessDTO, error) {
	policy, err := s.repository.GetBillPrintPolicy(ctx, c.BranchID)
	if err != nil {
		return nil, err
	}
	counts, err := s.repository.GetReadiness(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	return &CheckReadinessDTO{
		CheckDTO:  c.ToDTO(),
		Readiness: deriveReadiness(counts, policy),
		Counts:    counts,
	}, nil
}
