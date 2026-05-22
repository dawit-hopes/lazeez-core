package payment

import (
	"encoding/json"
	"fmt"
)

type PaymentPayload struct {
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	PhoneNumber   string `json:"phone_number"`
	TxRef         string `json:"tx_ref"`
	CallbackURL   string `json:"callback_url"`
	ReturnURL     string `json:"return_url"`
	Customization any    `json:"customization"`
	Meta          any    `json:"meta"`
}

type initializeAPIResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

type VerificationResponse struct {
	Message string           `json:"message"`
	Status  string           `json:"status"`
	Data    VerificationData `json:"data"`
}

type VerificationData struct {
	Amount   flexibleDecimal `json:"amount"`
	Currency string          `json:"currency"`
	Status   string          `json:"status"`
	TxRef    string          `json:"tx_ref"`
}

// flexibleDecimal accepts Chapa amount values as JSON strings or numbers.
type flexibleDecimal string

func (f *flexibleDecimal) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}

	switch data[0] {
	case '"':
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*f = flexibleDecimal(value)
		return nil
	default:
		var value json.Number
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*f = flexibleDecimal(value.String())
		return nil
	}
}

func (f flexibleDecimal) String() string {
	return string(f)
}

func (f flexibleDecimal) Float64() (float64, error) {
	if f == "" {
		return 0, fmt.Errorf("empty amount")
	}
	return json.Number(f.String()).Float64()
}

type VerificationResult struct {
	Status   string
	Amount   string
	Currency string
	TxRef    string
}
