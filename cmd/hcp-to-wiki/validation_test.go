package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

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

func TestParseDataMalformedRow(t *testing.T) {
	header := []string{"Source"}

	// This row is deliberately incomplete.
	malformed := []string{
		"Magazine",
		"1980-01",
		"p1",
		"Test Computer",
	}

	// Construct a complete row using the existing field definitions.
	valid := make([]string, adv_board+1)
	valid[adv_magazine] = "Magazine"
	valid[adv_yyyy_mm] = "1980-01"
	valid[adv_page_num] = "p1"
	valid[adv_system] = "Test Computer"
	valid[adv_price_inc_vat] = "£100"

	data := [][]string{
		header,
		malformed,
		valid,
	}

	adverts, _, _ := parseData(data)

	if len(adverts) != 1 {
		t.Fatalf("expected 1 advert, got %d", len(adverts))
	}

	if adverts[0].system != "Test Computer" {
		t.Errorf("unexpected system name: %s", adverts[0].system)
	}

	if adverts[0].price != 100 {
		t.Errorf("expected price 100, got %d", adverts[0].price)
	}
}

func TestParseDataEmptyRow(t *testing.T) {
	data := [][]string{
		{},
		{"Source"},
		{},
	}

	adverts, _, _ := parseData(data)

	if len(adverts) != 0 {
		t.Errorf("expected no adverts, got %d", len(adverts))
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer
	fn()

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}

	return string(output)
}

func TestOutputWikidataEmpty(t *testing.T) {
	output := captureStdout(t, func() {
		outputWikidata(
			make(map[string][]int),
			[]string{},
			(max_year+1)*4,
			-1,
		)
	})

	expected := []string{
		"__TOC__",
		`{| class="wikitable" style="width: 95%;"`,
		"! System",
		"|}",
		"[[Category:Classic Computing]]",
	}

	for _, text := range expected {
		if !strings.Contains(output, text) {
			t.Errorf("output missing %q", text)
		}
	}

	if strings.Contains(output, "== 2100 -") {
		t.Error("empty output should not contain a dated table")
	}
}
