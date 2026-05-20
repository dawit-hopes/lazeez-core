package common

import (
	"encoding/json"
	"testing"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	var v struct {
		N FlexInt `json:"n"`
	}
	for _, payload := range []string{`{"n":1}`, `{"n":1.0}`, `{"n":0}`} {
		if err := json.Unmarshal([]byte(payload), &v); err != nil {
			t.Fatalf("payload %s: %v", payload, err)
		}
	}
}
