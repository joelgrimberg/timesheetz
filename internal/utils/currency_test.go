package utils

import (
	"testing"
)

func TestFormatEuro(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{"whole number", 100.0, "€100,00"},
		{"with decimals", 100.50, "€100,50"},
		{"small amount", 0.01, "€0,01"},
		{"zero", 0.0, "€0,00"},
		{"large amount", 10000.99, "€10000,99"},
		{"negative", -50.25, "€-50,25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEuro(tt.amount)
			if result != tt.expected {
				t.Errorf("FormatEuro(%v) = %v, want %v", tt.amount, result, tt.expected)
			}
		})
	}
}

func TestParseEuro(t *testing.T) {
	tests := []struct {
		name      string
		euroStr   string
		expected  float64
		shouldErr bool
	}{
		{"with symbol and comma", "€100,50", 100.50, false},
		{"without symbol", "100,50", 100.50, false},
		{"whole number", "€100,00", 100.00, false},
		{"small amount", "€0,01", 0.01, false},
		{"with spaces", " €100,50 ", 100.50, false},
		{"negative", "€-50,25", -50.25, false},
		{"large amount", "€10000,99", 10000.99, false},
		{"invalid", "invalid", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEuro(tt.euroStr)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseEuro(%v) expected error, got nil", tt.euroStr)
				}
			} else {
				if err != nil {
					t.Errorf("ParseEuro(%v) unexpected error: %v", tt.euroStr, err)
				}
				if result != tt.expected {
					t.Errorf("ParseEuro(%v) = %v, want %v", tt.euroStr, result, tt.expected)
				}
			}
		})
	}
}

func TestFormatParseRoundtrip(t *testing.T) {
	tests := []float64{0.0, 0.01, 100.0, 100.50, 10000.99, -50.25}

	for _, amount := range tests {
		formatted := FormatEuro(amount)
		parsed, err := ParseEuro(formatted)
		if err != nil {
			t.Errorf("Roundtrip failed for %v: %v", amount, err)
		}
		if parsed != amount {
			t.Errorf("Roundtrip failed: %v -> %v -> %v", amount, formatted, parsed)
		}
	}
}
