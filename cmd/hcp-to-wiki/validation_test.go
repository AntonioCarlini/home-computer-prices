package main

import "testing"

func TestHandleYYYYMMInvalidLength(t *testing.T) {
	tests := []struct {
		name string
		date string
	}{
		{"empty", ""},
		{"one character", "1"},
		{"four characters", "2026"},
		{"five characters", "2026-"},
		{"six characters", "2026-1"},
		{"eight characters", "2026-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, month, err := handle_yyyy_mm(tt.date)

			if err == nil {
				t.Fatalf("expected error for date %q", tt.date)
			}

			if year != -1 || month != -1 {
				t.Errorf("expected year and month -1, got %d and %d", year, month)
			}
		})
	}
}

func TestHandleYYYYMMInvalidSeparator(t *testing.T) {
	_, _, err := handle_yyyy_mm("2026/01")

	if err == nil {
		t.Fatal("expected error for invalid separator")
	}
}

func TestHandleYYYYMMInvalidDigits(t *testing.T) {
	tests := []string{
		"abcd-01",
		"2026-ab",
	}

	for _, date := range tests {
		t.Run(date, func(t *testing.T) {
			_, _, err := handle_yyyy_mm(date)

			if err == nil {
				t.Errorf("expected error for date %q", date)
			}
		})
	}
}

func TestHandleYYYYMMInvalidMonth(t *testing.T) {
	tests := []string{
		"2026-00",
		"2026-13",
	}

	for _, date := range tests {
		t.Run(date, func(t *testing.T) {
			_, _, err := handle_yyyy_mm(date)

			if err == nil {
				t.Errorf("expected error for date %q", date)
			}
		})
	}
}

func TestHandleYYYYMMInvalidYear(t *testing.T) {
	tests := []string{
		"1944-12",
		"2100-01",
	}

	for _, date := range tests {
		t.Run(date, func(t *testing.T) {
			_, _, err := handle_yyyy_mm(date)

			if err == nil {
				t.Errorf("expected error for date %q", date)
			}
		})
	}
}

func TestHandleYYYYMMValid(t *testing.T) {
	tests := []struct {
		date  string
		year  int
		month int
	}{
		{"1945-01", 1945, 1},
		{"1980-06", 1980, 6},
		{"2099-12", 2099, 12},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			year, month, err := handle_yyyy_mm(tt.date)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if year != tt.year || month != tt.month {
				t.Errorf("expected %d/%d, got %d/%d",
					tt.year, tt.month, year, month)
			}
		})
	}
}
