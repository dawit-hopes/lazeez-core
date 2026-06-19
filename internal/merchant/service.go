package merchant

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/users"
	"strings"
)

type MerchantService interface {
	Create(ctx context.Context, req MerchantRequest) (*MerchantDTO, error)
	Get(ctx context.Context, id string) (*MerchantDTO, error)
	Update(ctx context.Context, id string, req MerchantRequest) error
	GetAll(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*MerchantDTO], error)
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	GetTaxCharges(ctx context.Context, merchantID, role, callerMerchantID string) (*TaxCharges, error)
	UpdateTaxCharges(ctx context.Context, merchantID string, req TaxChargesRequest, role, callerMerchantID string) (*TaxCharges, error)
}

type merchantService struct {
	merchantRepository MerchantRepository
	fileService        files.FileService
	logger             config.Logger
}

func NewMerchantService(merchantRepository MerchantRepository, fileService files.FileService, logger config.Logger) MerchantService {
	return &merchantService{
		merchantRepository: merchantRepository,
		fileService:        fileService,
		logger:             logger,
	}
}

func (s *merchantService) Create(ctx context.Context, req MerchantRequest) (*MerchantDTO, error) {
	merchant := Merchant{
		Name:       common.FormatText(req.Name),
		BranchType: req.BranchType,
		VatPercent: defaultVatPercent,
	}
	merchant.ID = common.GenerateUUID()
	err := s.merchantRepository.CheckExists(ctx, merchant.Name, "")
	if err != nil {
		s.logger.Error("Failed to check if merchant exists", "error", err)
		return nil, err
	}

	if req.Logo != nil {
		s.logger.Info("Uploading logo", "logo", req.LogoHeader.Filename)
		imageURL, err := s.fileService.UploadFile(ctx, &req.LogoHeader)
		if err != nil {
			s.logger.Error("Failed to upload logo", "error", err)
			return nil, err
		}

		s.logger.Info("image url", imageURL)
		merchant.Logo = imageURL
	}

	newMerchant, err := s.merchantRepository.Create(ctx, merchant)
	if err != nil {
		s.logger.Error("Failed to create merchant", "error", err)
		return nil, err
	}
	return newMerchant, nil
}

func (s *merchantService) Get(ctx context.Context, id string) (*MerchantDTO, error) {
	s.logger.Info("Getting merchant by ID", "id", id)
	merchant, err := s.merchantRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return nil, err
	}
	return merchant, nil
}

func (s *merchantService) Update(ctx context.Context, id string, req MerchantRequest) error {
	s.logger.Info("Updating merchant", "id", id, "name", req.Name)

	// Load existing merchant from repository
	existingMerchant, err := s.merchantRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get merchant by ID", "error", err)
		return err
	}

	// If name is provided and changed, ensure it's not taken by another merchant
	if req.Name != "" && !strings.EqualFold(req.Name, existingMerchant.Name) {
		if err := s.merchantRepository.CheckExists(ctx, req.Name, id); err != nil {
			s.logger.Error("Failed to check if merchant exists", "error", err)
			return err
		}
		existingMerchant.Name = common.FormatText(req.Name)
	}

	if req.Logo != nil {
		imageURL, err := s.fileService.UploadFile(ctx, &req.LogoHeader)
		if err != nil {
			s.logger.Error("Failed to upload logo", "error", err)
			return err
		}
		s.logger.Info("Uploaded logo", "imageURL", imageURL)
		existingMerchant.Logo = imageURL
	}

	if req.BranchType != "" {
		existingMerchant.BranchType = req.BranchType
	}

	err = s.merchantRepository.Update(ctx, existingMerchant.ToModel())
	if err != nil {
		s.logger.Error("Failed to update merchant", "error", err)
		return err
	}
	return nil
}

func (s *merchantService) Delete(ctx context.Context, id string) error {

	err := s.merchantRepository.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete merchant", "error", err)
		return err
	}
	return nil
}

func (s *merchantService) GetAll(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*MerchantDTO], error) {
	s.logger.Info("Getting all merchants")
	return s.merchantRepository.GetAll(ctx, filter)
}

func (s *merchantService) UnDelete(ctx context.Context, id string) error {
	s.logger.Info("Undeleting merchant", "id", id)
	err := s.merchantRepository.UnDelete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to undelete merchant", "error", err)
		return err
	}
	return nil
}

func canReadTaxCharges(role, callerMerchantID, resourceMerchantID string) bool {
	if users.IsSuperAdminRoleString(role) {
		return true
	}
	if users.CanManageMerchantMaster(role, callerMerchantID, resourceMerchantID) {
		return true
	}
	return users.IsBranchStaffRoleString(role) && callerMerchantID != "" && callerMerchantID == resourceMerchantID
}

func (s *merchantService) GetTaxCharges(ctx context.Context, merchantID, role, callerMerchantID string) (*TaxCharges, error) {
	if !canReadTaxCharges(role, callerMerchantID, merchantID) {
		return nil, common.ErrUnAuthorized
	}
	taxCharges, _, err := s.merchantRepository.GetTaxCharges(ctx, merchantID)
	if err != nil {
		s.logger.Error("Failed to get merchant tax charges", "error", err)
		return nil, err
	}
	return taxCharges, nil
}

func (s *merchantService) UpdateTaxCharges(ctx context.Context, merchantID string, req TaxChargesRequest, role, callerMerchantID string) (*TaxCharges, error) {
	if !users.CanManageMerchantMaster(role, callerMerchantID, merchantID) {
		return nil, common.ErrUnAuthorized
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	current, branchType, err := s.merchantRepository.GetTaxCharges(ctx, merchantID)
	if err != nil {
		s.logger.Error("Failed to load merchant tax charges", "error", err)
		return nil, err
	}
	if branchType != BranchTypeRestaurant {
		return nil, common.ErrInvalidRequest
	}

	vatPercent := current.VatPercent
	if req.VatPercent != nil {
		vatPercent = *req.VatPercent
	}

	var serviceCharge sql.NullFloat64
	if req.ServiceChargeSet {
		if req.ServiceChargePercent == nil {
			serviceCharge = sql.NullFloat64{}
		} else {
			serviceCharge = sql.NullFloat64{Float64: *req.ServiceChargePercent, Valid: true}
		}
	} else if current.ServiceChargePercent != nil {
		serviceCharge = sql.NullFloat64{Float64: *current.ServiceChargePercent, Valid: true}
	}

	updated, err := s.merchantRepository.UpdateTaxCharges(ctx, merchantID, vatPercent, serviceCharge)
	if err != nil {
		s.logger.Error("Failed to update merchant tax charges", "error", err)
		return nil, err
	}
	return updated, nil
}
