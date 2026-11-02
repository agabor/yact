package logic

import (
	"slices"
	"testing"
)

func TestParsePromptLinkTargetsDoubleBrackets(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{"double bracket", "see [[main.go]]", []string{"main.go"}},
		{"markdown link", "see [main.go](main.go)", []string{"main.go"}},
		{"no link", "plain text with main.go", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePromptLinkTargets(tt.prompt)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ParsePromptLinkTargets(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestParsePromptLinkTargetsMixed(t *testing.T) {
	got := ParsePromptLinkTargets("[[a.go]] and [b.go](b.go)")
	for _, want := range []string{"a.go", "b.go"} {
		if !slices.Contains(got, want) {
			t.Errorf("ParsePromptLinkTargets() = %v, missing %q", got, want)
		}
	}
}
