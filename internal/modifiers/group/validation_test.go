package group

import (
	"encoding/json"
	"testing"
)

func TestModifierGroupRequest_Validate(t *testing.T) {
	payload := `{"name":"Size","selection_type":"single","is_required":true,"min_selections":1,"max_selections":1,"options":[]}`
	var req ModifierGroupRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
