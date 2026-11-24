package main

import "testing"

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"first is smaller", 3, 5, 3},
		{"second is smaller", 5, 3, 3},
		{"equal values", 4, 4, 4},
		{"negative values", -5, -2, -5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := min(test.a, test.b)
			if got != test.expected {
				t.Errorf("expected %d, got %d", test.expected, got)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"first is larger", 5, 3, 5},
		{"second is larger", 3, 5, 5},
		{"equal values", 4, 4, 4},
		{"negative values", -2, -5, -2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := max(test.a, test.b)
			if got != test.expected {
				t.Errorf("expected %d, got %d", test.expected, got)
			}
		})
	}
}

func TestSliceContainsString(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		candidate string
		expected  bool
	}{
		{
			name:      "present",
			slice:     []string{"Apple", "Atari", "Commodore"},
			candidate: "Atari",
			expected:  true,
		},
		{
			name:      "absent",
			slice:     []string{"Apple", "Atari", "Commodore"},
			candidate: "Sinclair",
			expected:  false,
		},
		{
			name:      "empty slice",
			slice:     []string{},
			candidate: "Apple",
			expected:  false,
		},
		{
			name:      "nil slice",
			slice:     nil,
			candidate: "Apple",
			expected:  false,
		},
		{
			name:      "empty string present",
			slice:     []string{"Apple", ""},
			candidate: "",
			expected:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sliceContainsString(test.slice, test.candidate)
			if got != test.expected {
				t.Errorf("expected %t, got %t", test.expected, got)
			}
		})
	}
}

