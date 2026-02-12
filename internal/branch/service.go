package branch

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type BranchService interface {
	Create(ctx context.Context, req CreateBranchRequest) error
	Get(ctx context.Context, id string) (*BranchResponse, error)
	Update(ctx context.Context, id string, req UpdateBranchRequest) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*BranchResponse, error)
	GetAllByMerchantID(ctx context.Context, merchantID string) ([]*BranchResponse, error)
}

type branchService struct {
	branchRepository BranchRepository
	logger           config.Logger
}

func NewBranchService(branchRepository BranchRepository, logger config.Logger) BranchService {
	return &branchService{
		branchRepository: branchRepository,
		logger:           logger,
	}
}

func (s *branchService) Create(ctx context.Context, req CreateBranchRequest) error {
	err := s.branchRepository.CheckExists(ctx, req.MerchantID, req.BranchName, req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to check if branch exists", "error", err)
		return err
	}

	branch := Branch{
		MerchantID:  req.MerchantID,
		BranchName:  req.BranchName,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	}

	branch.ID = common.GenerateUUID()

	normalizedPhoneNumber, err := common.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		s.logger.Error("Failed to validate phone number", "error", err)
		return err
	}
	branch.PhoneNumber = normalizedPhoneNumber

	s.logger.Info("Creating branch", "branch", branch)

	err = s.branchRepository.Create(ctx, branch)
	if err != nil {
		s.logger.Error("Failed to create branch", "error", err)
		return err
	}

	return nil
}

func (s *branchService) Get(ctx context.Context, id string) (*BranchResponse, error) {
	s.logger.Info("Getting branch by ID", "id", id)
	branch, err := s.branchRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get branch", "error", err)
		return nil, err
	}

	branchDTO := branch.ToDTO()
	return &branchDTO, nil
}

func (s *branchService) Update(ctx context.Context, id string, req UpdateBranchRequest) error {
	s.logger.Info("Updating branch", "id", id)

	existingBranch, err := s.branchRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get branch", "error", err)
		return err
	}

	if req.BranchName != "" {
		existingBranch.BranchName = req.BranchName
	}
	if req.Address != "" {
		existingBranch.Address = req.Address
	}
	if req.PhoneNumber != "" {
		existingBranch.PhoneNumber = req.PhoneNumber
	}

	err = s.branchRepository.Update(ctx, existingBranch)
	if err != nil {
		s.logger.Error("Failed to update branch", "error", err)
		return err
	}

	return nil
}

func (s *branchService) Delete(ctx context.Context, id string) error {
	s.logger.Info("Deleting branch", "id", id)

	err := s.branchRepository.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete branch", "error", err)
		return err
	}

	return nil
}

func (s *branchService) GetAll(ctx context.Context) ([]*BranchResponse, error) {
	s.logger.Info("Getting all branches")
	branches, err := s.branchRepository.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all branches", "error", err)
		return nil, err
	}
	branchDTOs := make([]*BranchResponse, len(branches))
	for i, branch := range branches {
		result := branch.ToDTO()
		branchDTOs[i] = &result
	}
	return branchDTOs, nil
}

func (s *branchService) GetAllByMerchantID(ctx context.Context, merchantID string) ([]*BranchResponse, error) {
	s.logger.Info("Getting all branches by merchant ID", "merchantID", merchantID)
	branches, err := s.branchRepository.GetAllByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Error("Failed to get all branches by merchant ID", "error", err)
		return nil, err
	}
	branchDTOs := make([]*BranchResponse, len(branches))
	for i, branch := range branches {
		result := branch.ToDTO()
		branchDTOs[i] = &result
	}
	return branchDTOs, nil
}


