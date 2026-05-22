package option

import (
	"context"
	"lazeez-core/config"
)

type ModifierOptionService interface {
	Create(ctx context.Context, modifierOption ModifierOption) error
	Get(ctx context.Context, id string) (ModifierOption, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]ModifierOption, error)
	Update(ctx context.Context, modifierOption ModifierOption) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
}

type modifierOptionService struct {
	modifierOptionRepository ModifierOptionRepository
	logger                   config.Logger
}

func NewModifierOptionService(modifierOptionRepository ModifierOptionRepository, logger config.Logger) ModifierOptionService {
	return &modifierOptionService{modifierOptionRepository: modifierOptionRepository, logger: logger}
}

func (s *modifierOptionService) Create(ctx context.Context, modifierOption ModifierOption) error {
	return s.modifierOptionRepository.Create(ctx, modifierOption)
}

func (s *modifierOptionService) Get(ctx context.Context, id string) (ModifierOption, error) {
	return s.modifierOptionRepository.Get(ctx, id)
}

func (s *modifierOptionService) GetByIDs(ctx context.Context, ids []string) (map[string]ModifierOption, error) {
	return s.modifierOptionRepository.GetByIDs(ctx, ids)
}

func (s *modifierOptionService) Update(ctx context.Context, modifierOption ModifierOption) error {
	return s.modifierOptionRepository.Update(ctx, modifierOption)
}

func (s *modifierOptionService) Delete(ctx context.Context, id string) error {
	return s.modifierOptionRepository.Delete(ctx, id)
}

func (s *modifierOptionService) UnDelete(ctx context.Context, id string) error {
	return s.modifierOptionRepository.UnDelete(ctx, id)
}
