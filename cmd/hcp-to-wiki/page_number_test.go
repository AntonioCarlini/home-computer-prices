package main

import "testing"

func TestHandlePageNumberValid(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"p1", 1},
		{"p123", 123},
		{"p500", 500},
		{"p0", 0},
		{" p123 ", 123},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			page, err := handle_page_number(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if page != tt.want {
				t.Errorf("expected %d, got %d", tt.want, page)
			}
		})
	}
}

func TestHandlePageNumberInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"prefix only", "p"},
		{"missing prefix", "123"},
		{"uppercase prefix", "P123"},
		{"invalid digits", "pxyz"},
		{"negative", "p-1"},
		{"above maximum", "p501"},
		{"invalid suffix", "p123a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handle_page_number(tt.input)
			if err == nil {
				t.Fatalf("expected error for page number %q", tt.input)
			}
		})
	}
}