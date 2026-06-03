package menu

import (
	"encoding/json"
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/room"
	"testing"
)

func TestPublicMenuCatalogResponse_JSONEnvelope(t *testing.T) {
	catalog := PublicMenuCatalogResponse{
		Room:       &room.RoomResponseSimplified{RoomNumber: "101", Reference: "ref", Status: "occupied"},
		Guest:      &booking.GuestResponseSimplified{GuestName: "Jane Doe"},
		Meta:       common.PaginationMeta{TotalDocs: 3, Limit: 100, Page: 1},
		Categories: []*category.CategoryResponseSimplified{{Name: "Main"}},
	}

	dataBytes, err := json.Marshal(catalog)
	if err != nil {
		t.Fatalf("marshal catalog: %v", err)
	}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		t.Fatalf("unmarshal data envelope: %v", err)
	}

	if _, ok := data["meta"]; ok {
		t.Fatal("expected meta omitted from data payload")
	}
	if _, ok := data["guest"]; !ok {
		t.Fatal("expected guest inside data payload")
	}
	if _, ok := data["categories"]; !ok {
		t.Fatal("expected categories inside data payload")
	}

	root := struct {
		Data json.RawMessage        `json:"data"`
		Meta *common.PaginationMeta `json:"meta"`
	}{
		Data: dataBytes,
		Meta: &catalog.Meta,
	}
	rootBytes, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("marshal root response: %v", err)
	}

	var rootMap map[string]json.RawMessage
	if err := json.Unmarshal(rootBytes, &rootMap); err != nil {
		t.Fatalf("unmarshal root: %v", err)
	}
	if _, ok := rootMap["meta"]; !ok {
		t.Fatal("expected meta at response root")
	}

	var nestedData map[string]json.RawMessage
	if err := json.Unmarshal(rootMap["data"], &nestedData); err != nil {
		t.Fatalf("unmarshal nested data: %v", err)
	}
	if _, ok := nestedData["meta"]; ok {
		t.Fatal("expected no meta inside nested data")
	}
}
