/*
Copyright SUSE LLC.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
