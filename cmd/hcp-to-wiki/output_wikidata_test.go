package main

import (
	"strings"
	"testing"
)

func TestOutputWikidata(t *testing.T) {
	minDate := buildIndexFromYearAndQuarter(2000, 1)
	maxDate := buildIndexFromYearAndQuarter(2000, 1)

	systems := map[string][]int{
		"Test Computer": {125},
	}

	keys := []string{"Test Computer"}

	got := captureStdout(t, func() {
		outputWikidata(systems, keys, minDate, maxDate)
	})

	expectedStrings := []string{
		"__TOC__",
		"== 2000 - 2004 ==",
		"System",
		"Test Computer",
		"£125",
		"[[Category:Classic Computing]]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(got, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}
