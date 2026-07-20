package handler

import "testing"

func TestNormalizeProfileLink(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		valid bool
	}{
		{name: "clear", input: "", want: "", valid: true},
		{name: "https", input: "https://example.com/profile", want: "https://example.com/profile", valid: true},
		{name: "http", input: "http://example.com/profile", want: "http://example.com/profile", valid: true},
		{name: "missing scheme", input: "example.com/profile", want: "https://example.com/profile", valid: true},
		{name: "javascript", input: "javascript:alert(1)", valid: false},
		{name: "data", input: "data:text/html,unsafe", valid: false},
		{name: "ftp", input: "ftp://example.com/file", valid: false},
		{name: "missing host", input: "https://", valid: false},
		{name: "malformed", input: "http:example.com", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := normalizeProfileLink(tt.input)
			if valid != tt.valid || got != tt.want {
				t.Fatalf("normalizeProfileLink(%q) = (%q, %v), want (%q, %v)", tt.input, got, valid, tt.want, tt.valid)
			}
		})
	}
}
