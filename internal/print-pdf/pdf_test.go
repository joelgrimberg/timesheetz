package printPDF

import (
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove ANSI color codes",
			input:    "\x1b[31mRed text\x1b[0m",
			expected: "Red text",
		},
		{
			name:     "Remove ANSI escape sequences",
			input:    "\x1b[1;32mBold Green\x1b[0m",
			expected: "Bold Green",
		},
		{
			name:     "Replace box-drawing characters",
			input:    "┌─────┐\n│ Box │\n└─────┘",
			expected: "+-----+\n| Box |\n+-----+",
		},
		{
			name:     "Handle mixed content",
			input:    "\x1b[33m┌─Test─┐\x1b[0m",
			expected: "+-Test-+",
		},
		{
			name:     "Handle emoji replacement",
			input:    "Hello 💤 World",
			expected: "Hello    World",
		},
		{
			name:     "Plain text unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStripANSIPerformance(t *testing.T) {
	// Test that regex is pre-compiled by checking it doesn't panic
	// and that multiple calls work correctly
	input := strings.Repeat("\x1b[31mText\x1b[0m ", 100)

	for i := 0; i < 100; i++ {
		result := stripANSI(input)
		if result == "" {
			t.Error("stripANSI returned empty string")
		}
	}
}

func BenchmarkStripANSI(b *testing.B) {
	input := "\x1b[31m┌─────────────┐\x1b[0m\n\x1b[32m│ Test Data   │\x1b[0m\n\x1b[31m└─────────────┘\x1b[0m"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stripANSI(input)
	}
}
