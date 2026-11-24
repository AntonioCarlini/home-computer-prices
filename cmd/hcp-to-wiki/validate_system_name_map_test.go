package main

import (
	"io"
	"os"
	"testing"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer file.Close()

	originalStderr := os.Stderr
	os.Stderr = file
	defer func() {
		os.Stderr = originalStderr
	}()

	fn()

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek temporary file: %v", err)
	}

	output, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read temporary file: %v", err)
	}

	return string(output)
}

func TestValidateSystemNameMap(t *testing.T) {
	t.Run("reports missing variant", func(t *testing.T) {
		fullSystemNameMap := map[string][]string{
			"ZX Spectrum": {
				"ZX Spectrum (16K)",
				"ZX Spectrum",
			},
		}

		blankVariantSystems := map[string][]advertInfo{
			"ZX Spectrum": {
				{row: 10},
				{row: 12},
			},
		}

		got := captureStderr(t, func() {
			validateSystemNameMap(fullSystemNameMap, blankVariantSystems)
		})

		expected := "Suspect entry: ZX Spectrum   [ZX Spectrum (16K) ZX Spectrum] : REFS: 10 12 \n"

		if got != expected {
			t.Errorf("unexpected diagnostic output:\n got: %q\nwant: %q", got, expected)
		}
	})

	t.Run("does not report normal variants", func(t *testing.T) {
		fullSystemNameMap := map[string][]string{
			"ZX Spectrum": {
				"ZX Spectrum (16K)",
				"ZX Spectrum (48K)",
			},
			"Gemini Galaxy 2": {
				"Gemini Galaxy 2 (64K)",
			},
		}

		got := captureStderr(t, func() {
			validateSystemNameMap(fullSystemNameMap, nil)
		})

		if got != "" {
			t.Errorf("expected no diagnostic output, got: %q", got)
		}
	})
}
