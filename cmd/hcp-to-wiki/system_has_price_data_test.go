package main

import "testing"

func TestSystemHasPriceData(t *testing.T) {
	// Prices cover 2000 Q1 through 2001 Q4.
	minDate := buildIndexFromYearAndQuarter(2000, 1)
	maxDate := buildIndexFromYearAndQuarter(2001, 4)

	prices := []int{
		100, 0, 0, 0,
		0, 200, 0, 0,
	}

	tests := []struct {
		name      string
		startYear int
		endYear   int
		expected  bool
	}{
		{
			name:      "data in requested period",
			startYear: 2000,
			endYear:   2001,
			expected:  true,
		},
		{
			name:      "data only at beginning of period",
			startYear: 2000,
			endYear:   2000,
			expected:  true,
		},
		{
			name:      "data only at end of period",
			startYear: 2001,
			endYear:   2001,
			expected:  true,
		},
		{
			name:      "no data in requested period",
			startYear: 2002,
			endYear:   2003,
			expected:  false,
		},
		{
			name:      "requested period ends before available data",
			startYear: 1998,
			endYear:   1999,
			expected:  false,
		},
		{
			name:      "requested period overlaps available data",
			startYear: 1999,
			endYear:   2000,
			expected:  true,
		},
		{
			name:      "all prices zero",
			startYear: 2000,
			endYear:   2001,
			expected:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testPrices := prices
			if test.name == "all prices zero" {
				testPrices = make([]int, len(prices))
			}

			got := systemHasPriceData(
				test.startYear,
				test.endYear,
				minDate,
				maxDate,
				testPrices,
			)

			if got != test.expected {
				t.Errorf("expected %t, got %t", test.expected, got)
			}
		})
	}
}
