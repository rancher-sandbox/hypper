package eyecandy

import (
	"testing"
)

func TestEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "check if emoticons are output properly",
			input:    ":beer:",
			expected: "\U0001f37a",
		},
		{
			name:     "check if emoticons are output properly",
			input:    "\U0001f37a",
			expected: "\U0001f37a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ESPrint(false, tt.input)
			if tt.expected != s {
				t.Errorf("expected %s got %s", tt.expected, s)
			}
		})
	}
}
func TestRemovalOfEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string",
			input:    "this is a plain string",
			expected: "this is a plain string",
		},
		{
			name:     "plain utf-8 string",
			input:    "\u2000-\u3300",
			expected: "\u2000-\u3300",
		},
		{
			name:     "pure emoji string",
			input:    ":pizza::beer:",
			expected: "",
		},
		{
			name:     "mixed string",
			input:    ":pizza: beer",
			expected: " beer",
		},
		{
			name:     "weird edge cases",
			input:    ":pizza: :beer: :pizza beer:",
			expected: "  :pizza beer:",
		},
		{
			name:     "double colon",
			input:    ":: test",
			expected: ":: test",
		},
		{
			name:     "double colon emoji mixup",
			input:    "::pizza::",
			expected: "::",
		},
		{
			name:     "do not touch native utf-8 inside colons",
			input:    ":\u2000-\u3300:",
			expected: ":\u2000-\u3300:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := removeEmojiFromString(tt.input)
			if tt.expected != s {
				t.Errorf("expected %s got %s", tt.expected, s)
			}
		})
	}
}
