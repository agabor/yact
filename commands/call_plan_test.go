package commands

import (
	"testing"
)

func TestCombinePromptText(t *testing.T) {
	tests := []struct {
		name       string
		promptFile string
		cliPrompt  string
		want       string
	}{
		{name: "both inputs empty returns empty", promptFile: "", cliPrompt: "", want: ""},
		{name: "file text only", promptFile: "  file text \n", cliPrompt: "", want: "file text"},
		{name: "CLI text only", promptFile: " \n", cliPrompt: "cli text", want: "cli text"},
		{name: "both joined with a blank line after trimming", promptFile: " file \n", cliPrompt: "  cli  ", want: "file\n\ncli"},
		{name: "CLI text goes below the separator, read links stay above it", promptFile: "[a](a.go) kw\n-----\nplan body\n", cliPrompt: "cli text", want: "[a](a.go) kw\n\n-----\n\nplan body\n\ncli text"},
		{name: "read links kept above the separator when there is no CLI text", promptFile: "[a](a.go)\n-----\nplan body\n", cliPrompt: "", want: "[a](a.go)\n\n-----\n\nplan body"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := combinePromptText(tc.promptFile, tc.cliPrompt); got != tc.want {
				t.Errorf("combinePromptText(%q, %q) = %q, want %q", tc.promptFile, tc.cliPrompt, got, tc.want)
			}
		})
	}
}

func TestSentPromptText(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		want       string
	}{
		{
			name:       "read links above the separator are absent and keyword markers are removed from the body",
			promptText: "[a](a.go) kw\n-----\ndo ((x))\n",
			want:       "do x\n",
		},
		{
			name:       "prompt without separator has markers removed",
			promptText: "use ((error handling)) and (parens)",
			want:       "use error handling and (parens)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sentPromptText(tc.promptText); got != tc.want {
				t.Errorf("sentPromptText(%q) = %q, want %q", tc.promptText, got, tc.want)
			}
		})
	}
}
