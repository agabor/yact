package commands

import (
	"strings"
	"testing"
)

const fence = "```"

func codeFileBlock(path string, body string) string {
	return fence + "\n// " + path + "\n" + body + "\n" + fence + "\n"
}

func TestValidateCodeOnlyResponse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "clean code-only response passes",
			content: codeFileBlock("a.txt", "x"),
		},
		{
			name:    "whitespace-only text between blocks passes",
			content: codeFileBlock("a.txt", "x") + "   \n" + codeFileBlock("b.txt", "y"),
		},
		{
			name:    "non-blank free text fails",
			content: "intro\n" + codeFileBlock("a.txt", "x"),
			wantErr: "free text",
		},
		{
			name:    "incomplete code block fails",
			content: fence + "\n// a.txt\nx\n",
			wantErr: "incomplete",
		},
		{
			name:    "rejected file path fails",
			content: codeFileBlock("../escape.txt", "x"),
			wantErr: "rejected file path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCodeOnlyResponse(tc.content)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("validateCodeOnlyResponse() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateCodeOnlyResponse() error = nil, want error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("validateCodeOnlyResponse() error = %q, want it to contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}