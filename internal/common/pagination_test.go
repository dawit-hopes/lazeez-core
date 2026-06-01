package common

import (
	"net/http/httptest"
	"testing"
)

func TestParseFilterDefaultLimit(t *testing.T) {
	r := httptest.NewRequest("GET", "/items", nil)
	f := ParseFilter(r)
	if f.Limit != DefaultPageLimit {
		t.Fatalf("expected limit %d, got %d", DefaultPageLimit, f.Limit)
	}
	if f.Page != 1 {
		t.Fatalf("expected page 1, got %d", f.Page)
	}
}

func TestParseFilterExplicitLimit(t *testing.T) {
	r := httptest.NewRequest("GET", "/items?limit=50", nil)
	f := ParseFilter(r)
	if f.Limit != 50 {
		t.Fatalf("expected limit 50, got %d", f.Limit)
	}
}

func TestPageLimitFallback(t *testing.T) {
	page, limit := Filter{}.PageLimit()
	if page != 1 || limit != DefaultPageLimit {
		t.Fatalf("expected page 1 limit %d, got page %d limit %d", DefaultPageLimit, page, limit)
	}
}
