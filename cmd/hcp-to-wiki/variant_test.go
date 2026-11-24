
package main

import (
	"reflect"
	"testing"
)

func TestBuildFullSystemName(t *testing.T) {
	tests := []struct {
		name     string
		advert   advertInfo
		expected string
	}{
		{
			name: "blank variant",
			advert: advertInfo{
				system: "ZX Spectrum",
			},
			expected: "ZX Spectrum",
		},
		{
			name: "non-blank variant",
			advert: advertInfo{
				system:  "ZX Spectrum",
				variant: "16K",
			},
			expected: "ZX Spectrum (16K)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := buildFullSystemName(test.advert)
			if got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}

func TestBuildFullSystemNameMap(t *testing.T) {
	adverts := []advertInfo{
		{system: "ZX Spectrum", variant: "16K"},
		{system: "ZX Spectrum", variant: "48K"},
		{system: "ZX Spectrum", variant: "16K"},
		{system: "Computer A"},
		{system: "Computer B", variant: "Model 2"},
	}

	got := buildFullSystemNameMap(adverts)

	expected := map[string][]string{
		"ZX Spectrum": {
			"ZX Spectrum (16K)",
			"ZX Spectrum (48K)",
		},
		"Computer A": {
			"Computer A",
		},
		"Computer B": {
			"Computer B (Model 2)",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("unexpected full system name map:\nexpected: %#v\ngot:      %#v",
			expected, got)
	}
}

func TestBuildBlankVariantMap(t *testing.T) {
	adverts := []advertInfo{
		{row: 1, system: "ZX Spectrum", variant: "16K"},
		{row: 2, system: "ZX Spectrum"},
		{row: 3, system: "Computer A"},
		{row: 4, system: "Computer B", variant: "Model 2"},
	}

	got := buildBlankVariantMap(adverts)

        if len(got) != 3 {
            t.Fatalf("expected 3 systems in map, got %d", len(got))
        }

	expected := map[string][]advertInfo{
		"ZX Spectrum": {
			{row: 2, system: "ZX Spectrum"},
		},
		"Computer A": {
			{row: 3, system: "Computer A"},
		},
                "Computer B": {},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("unexpected blank variant map:\nexpected: %#v\ngot:      %#v",
			expected, got)
	}
}

func TestBuildAdvertSystemName(t *testing.T) {
	fullSystemNameMap := map[string][]string{
		"Computer A": {
			"Computer A",
		},
		"ZX Spectrum": {
			"ZX Spectrum (16K)",
			"ZX Spectrum (48K)",
		},
		"Computer B": {
			"Computer B (Model 2)",
		},
	}

	tests := []struct {
		name     string
		advert   advertInfo
		expected string
	}{
		{
			name: "single name",
			advert: advertInfo{
				system: "Computer A",
			},
			expected: "Computer A",
		},
		{
			name: "multiple variants",
			advert: advertInfo{
				system:  "ZX Spectrum",
				variant: "16K",
			},
			expected: "ZX Spectrum (16K)",
		},
		{
			name: "single non-blank variant",
			advert: advertInfo{
				system:  "Computer B",
				variant: "Model 2",
			},
			expected: "Computer B",
		},
		{
			name: "missing system",
			advert: advertInfo{
				system: "Unknown",
			},
			expected: "FAILED**FAILED",
		},
		{
			name: "empty name list",
			advert: advertInfo{
				system: "Empty",
			},
			expected: "BAD**BAD",
		},
	}

        fullSystemNameMap["Empty"] = []string{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := buildAdvertSystemName(test.advert, fullSystemNameMap)
			if got != test.expected {
				t.Errorf("expected %q, got %q", test.expected, got)
			}
		})
	}
}
