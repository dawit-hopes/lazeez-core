package group

import (
	"context"
	"lazeez-core/config"
)

type ModifierGroupService interface {
	Create(ctx context.Context, modifierGroup ModifierGroup) error
	Get(ctx context.Context, id string) (ModifierGroup, error)
	Update(ctx context.Context, modifierGroup ModifierGroup) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
}

type modifierGroupService struct {
	modifierGroupRepository ModifierGroupRepository
	logger                  config.Logger
}

func NewModifierGroupService(modifierGroupRepository ModifierGroupRepository, logger config.Logger) ModifierGroupService {
	return &modifierGroupService{modifierGroupRepository: modifierGroupRepository, logger: logger}
}

func (s *modifierGroupService) Create(ctx context.Context, modifierGroup ModifierGroup) error {
	return s.modifierGroupRepository.Create(ctx, modifierGroup)
}

func (s *modifierGroupService) Get(ctx context.Context, id string) (ModifierGroup, error) {
	return s.modifierGroupRepository.Get(ctx, id)
}

func (s *modifierGroupService) Update(ctx context.Context, modifierGroup ModifierGroup) error {
	return s.modifierGroupRepository.Update(ctx, modifierGroup)
}

func (s *modifierGroupService) Delete(ctx context.Context, id string) error {
	return s.modifierGroupRepository.Delete(ctx, id)
}

func (s *modifierGroupService) UnDelete(ctx context.Context, id string) error {
	return s.modifierGroupRepository.UnDelete(ctx, id)
}
