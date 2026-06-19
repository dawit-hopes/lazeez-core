package merchant

import (
	"database/sql"
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const defaultVatPercent = 15

type TaxCharges struct {
	VatPercent           float64  `json:"vat_percent"`
	ServiceChargePercent *float64 `json:"service_charge_percent,omitempty"`
}

type TaxChargesRequest struct {
	VatPercent           *float64 `json:"vat_percent"`
	ServiceChargePercent *float64 `json:"service_charge_percent"`
	ServiceChargeSet     bool     `json:"-"`
}

func (r *TaxChargesRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["vat_percent"]; ok {
		var f float64
		if err := json.Unmarshal(v, &f); err != nil {
			return err
		}
		r.VatPercent = &f
	}
	if v, ok := raw["service_charge_percent"]; ok {
		r.ServiceChargeSet = true
		if string(v) == "null" {
			r.ServiceChargePercent = nil
			return nil
		}
		var f float64
		if err := json.Unmarshal(v, &f); err != nil {
			return err
		}
		r.ServiceChargePercent = &f
	}
	return nil
}

func (r *TaxChargesRequest) Validate() error {
	if r.VatPercent == nil && !r.ServiceChargeSet {
		return validation.NewError("tax_charges", "at least one of vat_percent or service_charge_percent is required")
	}
	if r.VatPercent != nil {
		if err := validation.Validate(*r.VatPercent,
			validation.Min(0.0).Error("vat_percent must be at least 0"),
			validation.Max(100.0).Error("vat_percent must be at most 100"),
		); err != nil {
			return err
		}
	}
	if r.ServiceChargeSet && r.ServiceChargePercent != nil {
		if err := validation.Validate(*r.ServiceChargePercent,
			validation.Min(0.0).Error("service_charge_percent must be at least 0"),
			validation.Max(100.0).Error("service_charge_percent must be at most 100"),
		); err != nil {
			return err
		}
	}
	return nil
}

func taxChargesFromModel(vatPercent float64, serviceCharge sql.NullFloat64) *TaxCharges {
	tc := &TaxCharges{VatPercent: vatPercent}
	if serviceCharge.Valid {
		v := serviceCharge.Float64
		tc.ServiceChargePercent = &v
	}
	return tc
}

func applyTaxChargesToModel(merchant *Merchant, vatPercent float64, serviceCharge sql.NullFloat64) {
	merchant.VatPercent = vatPercent
	merchant.ServiceChargePercent = serviceCharge
}
