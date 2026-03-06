package scheduler

import (
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		expected string
	}{
		{
			name:     "No truncation needed",
			input:    "Hello World",
			maxRunes: 20,
			expected: "Hello World",
		},
		{
			name:     "Simple truncation",
			input:    "Hello World",
			maxRunes: 5,
			expected: "Hello...",
		},
		{
			name:     "UTF-8 truncation",
			input:    "你好世界",
			maxRunes: 2,
			expected: "你好...",
		},
		{
			name:     "Balanced Markdown bold",
			input:    "This is **bold** text",
			maxRunes: 15,
			expected: "This is **bold**...",
		},
		{
			name:     "Unbalanced Markdown bold - repair",
			input:    "This is **bold** text",
			maxRunes: 12, // "This is **bo"
			expected: "This is **bo**...",
		},
		{
			name:     "Broken Markdown tag at boundary - remove",
			input:    "This is **bold** text",
			maxRunes: 10, // "This is **" -> ends with **, removed
			expected: "This is...",
		},
		{
			name:     "Single star at boundary - remove",
			input:    "This is *bold* text",
			maxRunes: 9, // "This is *"
			expected: "This is...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxRunes)
			if got != tt.expected {
				t.Errorf("TruncateString() = %q, want %q", got, tt.expected)
			}
		})
	}
}
