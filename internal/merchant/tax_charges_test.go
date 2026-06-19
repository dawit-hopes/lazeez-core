package merchant

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestTaxChargesRequest_Validate(t *testing.T) {
	vat := 15.0
	service := 5.0

	tests := []struct {
		name    string
		req     TaxChargesRequest
		wantErr bool
	}{
		{
			name:    "empty request",
			req:     TaxChargesRequest{},
			wantErr: true,
		},
		{
			name:    "valid vat only",
			req:     TaxChargesRequest{VatPercent: &vat},
			wantErr: false,
		},
		{
			name:    "valid service charge only",
			req:     TaxChargesRequest{ServiceChargeSet: true, ServiceChargePercent: &service},
			wantErr: false,
		},
		{
			name: "vat out of range",
			req: TaxChargesRequest{
				VatPercent: ptrFloat64(101),
			},
			wantErr: true,
		},
		{
			name: "service charge out of range",
			req: TaxChargesRequest{
				ServiceChargeSet:     true,
				ServiceChargePercent: ptrFloat64(-1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxChargesRequest_UnmarshalJSON(t *testing.T) {
	var req TaxChargesRequest
	if err := json.Unmarshal([]byte(`{"vat_percent":15,"service_charge_percent":5}`), &req); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if req.VatPercent == nil || *req.VatPercent != 15 {
		t.Fatalf("unexpected vat_percent: %#v", req.VatPercent)
	}
	if !req.ServiceChargeSet || req.ServiceChargePercent == nil || *req.ServiceChargePercent != 5 {
		t.Fatalf("unexpected service_charge_percent: %#v", req.ServiceChargePercent)
	}

	var clearReq TaxChargesRequest
	if err := json.Unmarshal([]byte(`{"service_charge_percent":null}`), &clearReq); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if !clearReq.ServiceChargeSet || clearReq.ServiceChargePercent != nil {
		t.Fatalf("expected cleared service charge, got %#v", clearReq.ServiceChargePercent)
	}
}

func TestTaxChargesFromModel(t *testing.T) {
	withService := taxChargesFromModel(15, sql.NullFloat64{Float64: 5, Valid: true})
	if withService.VatPercent != 15 || withService.ServiceChargePercent == nil || *withService.ServiceChargePercent != 5 {
		t.Fatalf("unexpected tax charges: %#v", withService)
	}

	withoutService := taxChargesFromModel(15, sql.NullFloat64{})
	if withoutService.ServiceChargePercent != nil {
		t.Fatalf("expected nil service charge, got %#v", withoutService.ServiceChargePercent)
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}
