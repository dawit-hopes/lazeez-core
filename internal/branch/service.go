package branch

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type BranchService interface {
	Create(ctx context.Context, req CreateBranchRequest) (Branch, error)
	Get(ctx context.Context, id string) (Branch, error)
	Update(ctx context.Context, req UpdateBranchRequest) (Branch, error)
	Delete(ctx context.Context, id string) error
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

func (s *branchService) Create(ctx context.Context, req CreateBranchRequest) (Branch, error) {
	branch := Branch{
		MerchantID:  req.MerchantID,
		BranchName:  req.BranchName,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	}

	branch.ID = common.GenerateUUID()

	s.logger.Info("Creating branch", "branch", branch)

	createdBranch, err := s.branchRepository.Create(ctx, branch)
	if err != nil {
		s.logger.Error("Failed to create branch", "error", err)
		return branch, err
	}

	return createdBranch, nil
}

func (s *branchService) Get(ctx context.Context, id string) (Branch, error) {
	s.logger.Info("Getting branch by ID", "id", id)
	branch, err := s.branchRepository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get branch", "error", err)
		return Branch{}, err
	}
	return branch, nil
}

func (s *branchService) Update(ctx context.Context, req UpdateBranchRequest) (Branch, error) {
	s.logger.Info("Updating branch", "id", req.ID)

	existingBranch, err := s.branchRepository.Get(ctx, req.ID)
	if err != nil {
		s.logger.Error("Failed to get branch", "error", err)
		return Branch{}, err
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

	updatedBranch, err := s.branchRepository.Update(ctx, existingBranch)
	if err != nil {
		s.logger.Error("Failed to update branch", "error", err)
		return Branch{}, err
	}

	return updatedBranch, nil
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
