package common

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt unmarshals JSON numbers into int, accepting integer literals and
// whole-number floats (e.g. 1.0) produced by JavaScript clients.
type FlexInt int

func (n *FlexInt) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*n = 0
		return nil
	}

	var asInt int
	if err := json.Unmarshal(b, &asInt); err == nil {
		*n = FlexInt(asInt)
		return nil
	}

	var asFloat float64
	if err := json.Unmarshal(b, &asFloat); err != nil {
		return err
	}
	if asFloat != float64(int64(asFloat)) {
		return fmt.Errorf("expected whole number, got %v", asFloat)
	}
	*n = FlexInt(int(asFloat))
	return nil
}

func (n FlexInt) Int() int {
	return int(n)
}

func (n FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(n))
}

// ParseFlexInt parses form/query string values into FlexInt.
func ParseFlexInt(s string) (FlexInt, error) {
	if s == "" {
		return 0, nil
	}
	if i, err := strconv.Atoi(s); err == nil {
		return FlexInt(i), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if f != float64(int64(f)) {
		return 0, fmt.Errorf("expected whole number, got %v", f)
	}
	return FlexInt(int(f)), nil
}
