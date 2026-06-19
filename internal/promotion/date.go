package promotion

import (
	"fmt"
	"strings"
	"time"

	"lazeez-core/internal/common"
)

func parseFormDate(raw string) (common.Date, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return common.Date{}, nil
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		time.RFC3339Nano,
		"01/02/2006",
		"1/2/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return common.Date(t), nil
		}
	}
	return common.Date{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD, RFC3339, or MM/DD/YYYY", raw)
}
