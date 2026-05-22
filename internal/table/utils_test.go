package table

import "testing"

func TestBuildTableMenuURL(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		reference string
		want      string
	}{
		{
			name:      "base with trailing slash",
			baseURL:   "http://localhost:8082/",
			reference: "abc123",
			want:      "http://localhost:8082/abc123",
		},
		{
			name:      "legacy table prefix removed",
			baseURL:   "https://menu.example.com/table/",
			reference: "abc123",
			want:      "https://menu.example.com/abc123",
		},
		{
			name:      "reference without leading slash",
			baseURL:   "https://menu.example.com",
			reference: "/abc123",
			want:      "https://menu.example.com/abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTableMenuURL(tt.baseURL, tt.reference)
			if got != tt.want {
				t.Fatalf("buildTableMenuURL(%q, %q) = %q, want %q", tt.baseURL, tt.reference, got, tt.want)
			}
		})
	}
}
