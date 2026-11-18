package commands

import (
	"testing"
)

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
