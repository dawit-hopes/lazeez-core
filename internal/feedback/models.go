package feedback

import "lazeez-core/internal/common"


type FeedbackBase struct {
	common.Base
	HotelID     string `json:"hotel_id" db:"hotel_id"`
	BookingID   string `json:"booking_id" db:"booking_id"`
	GuestName   string `json:"guest_name" db:"guest_name"`
	Rating      int    `json:"rating" db:"rating"` // 1–5 overall
	Comment     string `json:"comment" db:"comment"`
	IsAnonymous bool   `json:"is_anonymous" db:"is_anonymous"`
}


type MealFeedback struct {
    FeedbackBase

    // What was ordered
    OrderID     string `json:"order_id" db:"order_id"`
    OrderItemID string `json:"order_item_id" db:"order_item_id"`
    MenuItemID  string `json:"menu_item_id" db:"menu_item_id"`
    MenuItemName string `json:"menu_item_name" db:"menu_item_name"`

    // Sub-ratings specific to food (1-5 each)
    TasteRating        int `json:"taste_rating" db:"taste_rating"`
    PresentationRating int `json:"presentation_rating" db:"presentation_rating"`
    TemperatureRating  int `json:"temperature_rating" db:"temperature_rating"`

    // Binary signals — quick taps for guest
    PortionSize    PortionSize `json:"portion_size" db:"portion_size"`
    DeliveredOnTime bool      `json:"delivered_on_time" db:"delivered_on_time"`
}

type PortionSize string

const (
    PortionTooSmall PortionSize = "too_small"
    PortionJustRight PortionSize = "just_right"
    PortionTooLarge  PortionSize = "too_large"
)


type OverallFeedback struct {
    FeedbackBase

    // Sub-ratings per category (1-5 each)
    RoomRating        int `json:"room_rating" db:"room_rating"`
    CleanlinessRating int `json:"cleanliness_rating" db:"cleanliness_rating"`
    StaffRating       int `json:"staff_rating" db:"staff_rating"`
    FoodRating        int `json:"food_rating" db:"food_rating"`
    ValueRating       int `json:"value_rating" db:"value_rating"`

    // Simple signals
    WouldRecommend bool `json:"would_recommend" db:"would_recommend"`
    // Net Promoter Score 0-10
    // "How likely are you to recommend us?"
    NPSScore       int  `json:"nps_score" db:"nps_score"`
}