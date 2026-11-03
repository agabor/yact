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
	tests := []struct {
		name    string
		content string
		keyword string
		want    bool
	}{
		{"exact match", "func HandleError() error", "HandleError", true},
		{"different case", "func HandleError() error", "handleerror", true},
		{"absent keyword", "func HandleError() error", "missing", false},
		{"plan matches PlanStep", "type PlanStep struct", "plan", true},
		{"plan matches plan_gate", "plan_gate.go", "plan", true},
		{"plan not inside explanation", "explanation of the design", "plan", false},
		{"words in order", "countInputLines", "input lines", true},
		{"words out of order", "lines input", "input lines", false},
		{"keyword with underscores", "max_input_lines", "max_input_lines", true},
		{"separator style differs", "max_input_lines", "input-lines", true},
		{"empty keyword", "anything", "", false},
		{"punctuation only keyword", "anything", "--", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsKeyword(tt.content, tt.keyword); got != tt.want {
				t.Errorf("ContainsKeyword(%q, %q) = %v, want %v", tt.content, tt.keyword, got, tt.want)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"camelCase", "handleError", []string{"handle", "error"}},
		{"acronym followed by word", "HTTPServer", []string{"http", "server"}},
		{"snake_case", "max_input_lines", []string{"max", "input", "lines"}},
		{"kebab-case", "input-lines", []string{"input", "lines"}},
		{"digits stay with their word", "v2Server", []string{"v2", "server"}},
		{"mixed separators", "  Plan--Step_v2.go ", []string{"plan", "step", "v2", "go"}},
		{"empty", "", nil},
		{"only separators", "__--", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitWords(tt.text); !slices.Equal(got, tt.want) {
				t.Errorf("splitWords(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
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
