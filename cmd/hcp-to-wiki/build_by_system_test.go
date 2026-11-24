package main

import "testing"

func TestBuildBySystem(t *testing.T) {
	adverts := []advertInfo{
		{
			year: 1980, month: 1,
			system: "Computer A", price: 100,
		},
		{
			year: 1980, month: 2,
			system: "Computer A", price: 80,
		},
		{
			year: 1980, month: 4,
			system: "Computer A", price: 120,
		},
		{
			year: 1980, month: 1,
			system: "ZX Spectrum", variant: "16K", price: 150,
		},
		{
			year: 1980, month: 1,
			system: "ZX Spectrum", variant: "48K", price: 200,
		},
		{
			year: 1980, month: 2,
			system: "ZX Spectrum", variant: "16K", price: 140,
		},
	}

	minDate := buildIndexFromYearAndQuarter(1980, 1)
	maxDate := buildIndexFromYearAndQuarter(1980, 2)

	blankVariantSystems := make(map[string][]advertInfo)

	result := buildBySystem(adverts, blankVariantSystems, minDate, maxDate)

	// Computer A has only one variant (blank), so no suffix is needed.
	computerA, ok := result["Computer A"]
	if !ok {
		t.Fatal("Computer A missing from result")
	}

	if len(computerA) != 2 {
		t.Fatalf("expected 2 price entries for Computer A, got %d",
			len(computerA))
	}

	if computerA[0] != 80 {
		t.Errorf("expected Computer A Q1 price 80, got %d", computerA[0])
	}
	if computerA[1] != 120 {
		t.Errorf("expected Computer A Q2 price 120, got %d", computerA[1])
	}

	// ZX Spectrum has two variants, so both should have separate names.
        if _, ok := result["ZX Spectrum (16K)"]; !ok {
            t.Fatal("ZX Spectrum (16K) missing from result")
        }

	spectrum48, ok := result["ZX Spectrum (48K)"]
	if !ok {
		t.Fatal("ZX Spectrum (48K) missing from result")
	}

        if got := result["ZX Spectrum (16K)"][0]; got != 140 {
            t.Errorf("expected ZX Spectrum (16K) Q1 price 140, got %d", got)
        }

        if got := result["ZX Spectrum (16K)"][1]; got != 0 {
            t.Errorf("expected ZX Spectrum (16K) Q2 price 0, got %d", got)
        }
	if spectrum48[1] != 0 {
		t.Errorf("expected ZX Spectrum (48K) Q2 price 0, got %d",
			spectrum48[1])
	}
}
