package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// Date represents a calendar date. JSON accepts "2006-01-02" or RFC3339 timestamps.
type Date time.Time

func (d *Date) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*d = Date{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		*d = Date{}
		return nil
	}

	for _, layout := range []string{dateLayout, time.RFC3339, time.RFC3339Nano} {
		if t, err := time.Parse(layout, s); err == nil {
			*d = Date(t)
			return nil
		}
	}
	return fmt.Errorf("invalid date %q: expected YYYY-MM-DD or RFC3339", s)
}

func (d Date) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(time.Time(d).Format(dateLayout))
}

func (d Date) Time() time.Time {
	return time.Time(d)
}

func (d Date) IsZero() bool {
	return time.Time(d).IsZero()
}
