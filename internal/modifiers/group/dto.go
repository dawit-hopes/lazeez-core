package group

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/modifiers/option"
)

type ModifierGroupDTO struct {
	ID            string                     `json:"id"`
	Name          string                     `json:"name"`
	SelectionType SelectionType              `json:"selection_type"`
	IsRequired    bool                       `json:"is_required"`
	MinSelections common.FlexInt               `json:"min_selections,omitempty"`
	MaxSelections common.FlexInt               `json:"max_selections,omitempty"`
	Options       []option.ModifierOptionDTO `json:"options,omitempty"`
}

type ModifierGroupRequest struct {
	Name          string                         `json:"name"`
	SelectionType SelectionType                  `json:"selection_type"`
	IsRequired    bool                           `json:"is_required"`
	MinSelections common.FlexInt               `json:"min_selections,omitempty"`
	MaxSelections common.FlexInt               `json:"max_selections,omitempty"`
	Options       []option.ModifierOptionRequest `json:"options,omitempty"`
}
