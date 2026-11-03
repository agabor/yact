package logic

import (
	"slices"
	"testing"
)

func TestParsePromptKeywords(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{"phrase with spaces", "find ((error handling)) please", []string{"error handling"}},
		{"duplicates removed", "((a)) and ((a)) again", []string{"a"}},
		{"empty phrase ignored", "nothing (()) here", nil},
		{"whitespace only ignored", "nothing (( )) here", nil},
		{"padding trimmed", "((  padded  ))", []string{"padded"}},
		{"order of first appearance", "((x)) ((y)) ((x))", []string{"x", "y"}},
		{"no phrases", "plain text (with parentheses)", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePromptKeywords(tt.prompt)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ParsePromptKeywords(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestContainsKeyword(t *testing.T) {
	content := "func HandleError() error"

	if !ContainsKeyword(content, "HandleError") {
		t.Error("expected match for present keyword")
	}
	if ContainsKeyword(content, "handleerror") {
		t.Error("expected no match for different case")
	}
	if ContainsKeyword(content, "missing") {
		t.Error("expected no match for absent keyword")
	}
}

func TestUniquePaths(t *testing.T) {
	input := []string{"b.go", "a.go", "b.go", "c.go", "a.go"}
	want := []string{"b.go", "a.go", "c.go"}

	got := UniquePaths(input)
	if !slices.Equal(got, want) {
		t.Errorf("UniquePaths(%v) = %v, want %v", input, got, want)
	}
	if !slices.Equal(input, []string{"b.go", "a.go", "b.go", "c.go", "a.go"}) {
		t.Errorf("UniquePaths mutated its argument: %v", input)
	}
}

func TestStripKeywordMarkers(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{"single marker", "find ((error handling)) now", "find error handling now"},
		{"several markers in one line", "((a)) and ((b))", "a and b"},
		{"text without markers", "plain text", "plain text"},
		{"single parentheses stay", "f(x) and (y)", "f(x) and (y)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripKeywordMarkers(tt.prompt); got != tt.want {
				t.Errorf("StripKeywordMarkers(%q) = %q, want %q", tt.prompt, got, tt.want)
			}
		})
	}
}