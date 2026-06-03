package menu

import (
	"lazeez-core/internal/branch"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/ingredient"
	"lazeez-core/internal/merchant"
	"lazeez-core/internal/modifiers/group"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/room"
	"lazeez-core/internal/table"
	"mime/multipart"

	"github.com/lib/pq"
)

type MenuRequest struct {
	Name            string                       `json:"name"`
	ImageHeader     multipart.FileHeader         `json:"-"`
	Image           multipart.File               `json:"image"`
	Description     string                       `json:"description"`
	Price           float64                      `json:"price"`
	Ingredients     []string                     `json:"ingredients"`
	CategoryID      string                       `json:"category_id"`
	BranchID        string                       `json:"branch_id"`
	MerchantID      string                       `json:"merchant_id"`
	IsFasting       *bool                        `json:"is_fasting"`
	PreparationTime float64                      `json:"preparation_time"`
	IsAvailable     *bool                        `json:"is_available"`
	Modifiers       []group.ModifierGroupRequest `json:"modifier_groups"`
	ModifiersSet    bool                         `json:"-"`
}

type MenuDTO struct {
	common.BaseDTO
	Name            string                      `json:"name"`
	Image           string                      `json:"image"`
	Description     string                      `json:"description"`
	Price           float64                     `json:"price"`
	CategoryID      string                      `json:"category_id"`
	Category        *category.CategoryDTO       `json:"category"`
	Ingredients     []*ingredient.IngredientDTO `json:"ingredients"`
	BranchID        string                      `json:"branch_id"`
	MerchantID      string                      `json:"merchant_id"`
	IsMaster        bool                        `json:"is_master"`
	IsExcluded      bool                        `json:"is_excluded"`
	IsFasting       bool                        `json:"is_fasting"`
	IsAvailable     bool                        `json:"is_available"`
	PreparationTime float64                     `json:"preparation_time"`
	Modifiers       []group.ModifierGroupDTO    `json:"modifier_groups"`
}

func (m *MenuRequest) IsEmpty() bool {
	return m.Name == "" && m.Image == nil && m.Description == "" && m.Price == 0 && len(m.Ingredients) == 0 && m.CategoryID == "" && m.BranchID == "" && m.IsFasting == nil && m.IsAvailable == nil && m.PreparationTime == 0 && !m.ModifiersSet
}

func (m *MenuRequest) ToModel(isMaster bool) Menu {
	ingredients := make(pq.StringArray, len(m.Ingredients))
	copy(ingredients, m.Ingredients)
	branchID := common.ToNUllString(m.BranchID)
	merchantID := common.ToNUllString(m.MerchantID)
	if isMaster {
		branchID = common.ToNUllString("")
	}
	isFasting := false
	if m.IsFasting != nil {
		isFasting = *m.IsFasting
	}
	isAvailable := true
	if m.IsAvailable != nil {
		isAvailable = *m.IsAvailable
	}
	return Menu{
		Name:            common.FormatText(m.Name),
		Description:     m.Description,
		Price:           m.Price,
		Ingredients:     ingredients,
		CategoryID:      common.ParseStringToUUID(m.CategoryID),
		BranchID:        branchID,
		MerchantID:      merchantID,
		IsFasting:       isFasting,
		IsAvailable:     isAvailable,
		PreparationTime: m.PreparationTime,
	}
}

// ModifierOptionPublic is a guest-facing modifier option (no audit fields).
type ModifierOptionPublic struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	PriceAdjustment float64 `json:"price_adjustment"`
	IsDefault       bool    `json:"is_default,omitempty"`
	IsAvailable     bool    `json:"is_available"`
}

// ModifierGroupPublic is a guest-facing modifier group (no audit fields).
type ModifierGroupPublic struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	SelectionType group.SelectionType    `json:"selection_type"`
	IsRequired    bool                   `json:"is_required"`
	MinSelections common.FlexInt         `json:"min_selections,omitempty"`
	MaxSelections common.FlexInt         `json:"max_selections,omitempty"`
	Options       []ModifierOptionPublic `json:"options,omitempty"`
}

// MenuDTOPublic is a single menu item for the guest-facing catalog.
type MenuDTOPublic struct {
	ID              string                                    `json:"id"`
	Name            string                                    `json:"name"`
	Image           string                                    `json:"image,omitempty"`
	Description     string                                    `json:"description,omitempty"`
	Price           float64                                   `json:"price"`
	Category        *category.CategoryResponseSimplified      `json:"category,omitempty"`
	Ingredients     []ingredient.IngredientResponseSimplified `json:"ingredients"`
	IsFasting       bool                                      `json:"is_fasting,omitempty"`
	IsAvailable     bool                                      `json:"is_available"`
	PreparationTime float64                                   `json:"preparation_time,omitempty"`
	Modifiers       []ModifierGroupPublic                     `json:"modifier_groups,omitempty"`
}

// PublicMenuCatalogResponse is returned when a guest scans a table or room QR (reference).
// Table or room (by endpoint), branch, and merchant are resolved once from the reference; menus are paginated.
type PublicMenuCatalogResponse struct {
	Table      *table.TableResponseSimplified         `json:"table,omitempty"`
	Room       *room.RoomResponseSimplified           `json:"room,omitempty"`
	Guest      *booking.GuestResponseSimplified       `json:"guest,omitempty"`
	Branch     *branch.BranchResponseSimplified       `json:"branch"`
	Merchant   *merchant.MerchantResponseSimplified   `json:"merchant"`
	Menus      []*MenuDTOPublic                       `json:"menus"`
	Meta       common.PaginationMeta                  `json:"-"`
	Categories []*category.CategoryResponseSimplified `json:"categories"`
}
