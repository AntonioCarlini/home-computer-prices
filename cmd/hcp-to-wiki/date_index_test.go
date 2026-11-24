package main

import "testing"

func TestBuildIndexFromYearAndQuarter(t *testing.T) {
	tests := []struct {
		year    int
		quarter int
		want    int
	}{
		{1945, 1, 1945 * 4},
		{1945, 2, 1945*4 + 1},
		{1945, 3, 1945*4 + 2},
		{1945, 4, 1945*4 + 3},
		{1980, 1, 1980 * 4},
		{1980, 4, 1980*4 + 3},
		{2099, 1, 2099 * 4},
		{2099, 4, 2099*4 + 3},
	}

	for _, tt := range tests {
		t.Run(
			string(rune('0'+tt.quarter)),
			func(t *testing.T) {
				got := buildIndexFromYearAndQuarter(tt.year, tt.quarter)
				if got != tt.want {
					t.Errorf("expected index %d, got %d", tt.want, got)
				}
			},
		)
	}
}

func TestBuildIndexFromAdvertInfo(t *testing.T) {
	tests := []struct {
		year  int
		month int
		want  int
	}{
		{1980, 1, 1980 * 4},
		{1980, 3, 1980 * 4},
		{1980, 4, 1980*4 + 1},
		{1980, 6, 1980*4 + 1},
		{1980, 7, 1980*4 + 2},
		{1980, 9, 1980*4 + 2},
		{1980, 10, 1980*4 + 3},
		{1980, 12, 1980*4 + 3},
	}

	for _, tt := range tests {
		advert := advertInfo{
			year:  tt.year,
			month: tt.month,
		}

		got := buildIndexFromAdvertInfo(advert)
		if got != tt.want {
			t.Errorf("%04d-%02d: expected index %d, got %d",
				tt.year, tt.month, tt.want, got)
		}
	}
}

func TestDecodeIndexByQuarter(t *testing.T) {
	tests := []struct {
		year    int
		quarter int
	}{
		{1945, 1},
		{1945, 4},
		{1980, 1},
		{1980, 2},
		{1980, 3},
		{1980, 4},
		{2099, 1},
		{2099, 4},
	}

	for _, tt := range tests {
		index := buildIndexFromYearAndQuarter(tt.year, tt.quarter)

		gotYear, gotQuarter := decodeIndexByQuarter(index)

		if gotYear != tt.year || gotQuarter != tt.quarter {
			t.Errorf("index %d: expected %04d Q%d, got %04d Q%d",
				index, tt.year, tt.quarter, gotYear, gotQuarter)
		}
	}
}
