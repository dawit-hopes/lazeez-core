package promotion

import (
	"context"

	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"
	"lazeez-core/internal/users"
)

type PromotionService interface {
	Create(ctx context.Context, req PromotionRequest, role string) (*PromotionDTO, error)
	Get(ctx context.Context, id, merchantID, role string) (*PromotionDTO, error)
	Update(ctx context.Context, id string, req PromotionRequest, role string) error
	Delete(ctx context.Context, id, merchantID, role string) error
	UnDelete(ctx context.Context, id, merchantID, role string) error
	List(ctx context.Context, filter common.Filter, merchantID, role string) (*common.PaginatedResponse[[]*PromotionDTO], error)
	ListActivePublicByReference(ctx context.Context, reference, referenceType string) ([]*PromotionPublicDTO, error)
}

type promotionService struct {
	promotionRepository PromotionRepository
	fileService         files.FileService
	logger              config.Logger
}

func NewPromotionService(
	promotionRepository PromotionRepository,
	fileService files.FileService,
	logger config.Logger,
) PromotionService {
	return &promotionService{
		promotionRepository: promotionRepository,
		fileService:         fileService,
		logger:              logger,
	}
}

func (s *promotionService) authorizeManage(role, userMerchantID, resourceMerchantID string) error {
	if users.CanManageMerchantMaster(role, userMerchantID, resourceMerchantID) {
		return nil
	}
	return common.ErrUnAuthorized
}

func (s *promotionService) Create(ctx context.Context, req PromotionRequest, role string) (*PromotionDTO, error) {
	if err := s.authorizeManage(role, req.MerchantID, req.MerchantID); err != nil {
		return nil, err
	}
	if req.MerchantID == "" {
		return nil, common.ErrBranchAdminMissingMerchant
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	bannerURL, err := s.fileService.UploadFile(ctx, &req.BannerImageHeader)
	if err != nil {
		s.logger.Error("failed to upload promotion banner", "error", err)
		return nil, err
	}

	promo := Promotion{
		MerchantID:  req.MerchantID,
		Title:       common.FormatText(req.Title),
		Description: req.Description,
		BannerImage: bannerURL,
		IsActive:    isActive,
		StartDate:   req.StartDate.Time(),
		EndDate:     req.EndDate.Time(),
	}
	promo.ID = common.GenerateUUID()

	if err := s.promotionRepository.Create(ctx, promo); err != nil {
		return nil, err
	}

	dto := promo.ToDTO()
	return &dto, nil
}

func (s *promotionService) Get(ctx context.Context, id, merchantID, role string) (*PromotionDTO, error) {
	promo, err := s.loadPromotionForRead(ctx, id, merchantID, role)
	if err != nil {
		return nil, err
	}
	dto := promo.ToDTO()
	return &dto, nil
}

func (s *promotionService) Update(ctx context.Context, id string, req PromotionRequest, role string) error {
	existing, err := s.promotionRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeManage(role, req.MerchantID, existing.MerchantID); err != nil {
		return err
	}

	if req.Title != "" {
		existing.Title = common.FormatText(req.Title)
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	start := existing.StartDate
	end := existing.EndDate
	if !req.StartDate.IsZero() {
		start = req.StartDate.Time()
		existing.StartDate = start
	}
	if !req.EndDate.IsZero() {
		end = req.EndDate.Time()
		existing.EndDate = end
	}
	if end.Before(start) {
		return common.ErrInvalidRequest
	}

	if req.BannerImage != nil {
		bannerURL, err := s.fileService.UploadFile(ctx, &req.BannerImageHeader)
		if err != nil {
			s.logger.Error("failed to upload promotion banner", "error", err)
			return err
		}
		existing.BannerImage = bannerURL
	}

	return s.promotionRepository.Update(ctx, existing)
}

func (s *promotionService) Delete(ctx context.Context, id, merchantID, role string) error {
	promo, err := s.loadPromotionForRead(ctx, id, merchantID, role)
	if err != nil {
		return err
	}
	if err := s.authorizeManage(role, merchantID, promo.MerchantID); err != nil {
		return err
	}
	return s.promotionRepository.Delete(ctx, id, promo.MerchantID)
}

func (s *promotionService) UnDelete(ctx context.Context, id, merchantID, role string) error {
	promo, err := s.promotionRepository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeManage(role, merchantID, promo.MerchantID); err != nil {
		return err
	}
	return s.promotionRepository.UnDelete(ctx, id, promo.MerchantID)
}

func (s *promotionService) List(ctx context.Context, filter common.Filter, merchantID, role string) (*common.PaginatedResponse[[]*PromotionDTO], error) {
	if merchantID == "" {
		return nil, common.ErrBranchAdminMissingMerchant
	}
	if !users.IsSuperAdminRoleString(role) && !users.CanManageMerchantMaster(role, merchantID, merchantID) {
		return nil, common.ErrUnAuthorized
	}

	result, err := s.promotionRepository.ListByMerchant(ctx, filter, merchantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*PromotionDTO, len(result.Data))
	for i, promo := range result.Data {
		dto := promo.ToDTO()
		dtos[i] = &dto
	}
	return &common.PaginatedResponse[[]*PromotionDTO]{
		Data: dtos,
		Meta: result.Meta,
	}, nil
}

func (s *promotionService) ListActivePublicByReference(ctx context.Context, reference, referenceType string) ([]*PromotionPublicDTO, error) {
	promos, err := s.promotionRepository.ListActivePublicByReference(ctx, reference, referenceType)
	if err != nil {
		return nil, err
	}

	dtos := make([]*PromotionPublicDTO, len(promos))
	for i, promo := range promos {
		dto := promo.ToPublicDTO()
		dtos[i] = &dto
	}
	return dtos, nil
}

func (s *promotionService) loadPromotionForRead(ctx context.Context, id, merchantID, role string) (Promotion, error) {
	if users.IsSuperAdminRoleString(role) && merchantID != "" {
		return s.promotionRepository.GetByMerchant(ctx, id, merchantID)
	}
	if merchantID == "" {
		return Promotion{}, common.ErrBranchAdminMissingMerchant
	}
	if !users.IsSuperAdminRoleString(role) && !users.CanManageMerchantMaster(role, merchantID, merchantID) {
		return Promotion{}, common.ErrUnAuthorized
	}
	return s.promotionRepository.GetByMerchant(ctx, id, merchantID)
}
