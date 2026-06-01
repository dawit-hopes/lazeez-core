package common

import "testing"

func TestBuildMenuURL(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		reference   string
		serviceType string
		want        string
	}{
		{
			name:        "table default base",
			base:        "http://localhost:8082",
			reference:   "7j23qTxXf2I",
			serviceType: "table",
			want:        "http://localhost:8082/tbl/7j23qTxXf2I",
		},
		{
			name:        "room default base",
			base:        "http://localhost:8082/",
			reference:   "7j23qTxXf2I",
			serviceType: "room",
			want:        "http://localhost:8082/rm/7j23qTxXf2I",
		},
		{
			name:        "base already has tbl suffix",
			base:        "https://menu.example.com/tbl",
			reference:   "abc",
			serviceType: "table",
			want:        "https://menu.example.com/tbl/abc",
		},
		{
			name:        "base already has rm suffix room",
			base:        "https://menu.example.com/rm",
			reference:   "abc",
			serviceType: "room",
			want:        "https://menu.example.com/rm/abc",
		},
		{
			name:        "strips wrong suffix when switching type",
			base:        "https://menu.example.com/tbl",
			reference:   "xyz",
			serviceType: "room",
			want:        "https://menu.example.com/rm/xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMenuURL(tt.base, tt.reference, tt.serviceType)
			if got != tt.want {
				t.Fatalf("buildMenuURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
