package main

import (
	"reflect"
	"testing"
)

func TestPreprocessSystemData(t *testing.T) {
	systems := map[string][]int{
		"Apple II":                  {100, 200},
		"Commodore PET":             {300, 400},
		"Exidy Sorcerer":            {500, 600},
		"Tandy TRS-80 Model 1":      {700, 800},
		"xApple II":                 {10, 20},
		"xCommodore PET":            {30, 40},
		"xExidy Sorcerer":           {50, 60},
		"xTandy TRS-80 Model 1":     {70, 80},
		"Science of Cambridge MK14": {900, 1000},
		"ZX Spectrum":               {1100, 1200},
	}

	got := preprocessSystemData(systems)

	expected := map[string][]int{
		"Apple II":      {100, 200},
		"Commodore PET": {300, 400},
		"Exidy Sorcerer": {500, 600},
		"Tandy TRS-80 Model 1": {700, 800},
		"MK14":          {900, 1000},
		"ZX Spectrum":   {1100, 1200},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("unexpected preprocessed data:\nexpected: %#v\ngot:      %#v",
			expected, got)
	}
}
