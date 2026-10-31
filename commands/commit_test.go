package commands

import (
	"errors"
	"testing"

	"yact/logic"
)

func TestPrintFreeText(t *testing.T) {
	elements := []logic.ParsedElement{
		logic.NewTextElement("intro"),
		logic.NewCodeFileElement(logic.CodeFile{Path: "a.go", Content: "x"}),
		logic.NewTextElement("   "),
		logic.NewRejectedFileElement(errors.New("bad path")),
		logic.NewTextElement("outro"),
	}

	got := captureStdout(t, func() {
		printFreeText(elements)
	})

	want := "intro\noutro\n"
	if got != want {
		t.Errorf("printFreeText() output = %q, want %q", got, want)
	}
}

func TestParseBufferElements(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		validateCode   bool
		wantErr        bool
		wantIncomplete bool
		wantOutput     string
	}{
		{
			name:         "clean code-only buffer parses with validation",
			content:      codeFileBlock("a.txt", "x"),
			validateCode: true,
		},
		{
			name:         "free text without validation is printed",
			content:      "intro\n" + codeFileBlock("a.txt", "x"),
			validateCode: false,
			wantOutput:   "intro\n",
		},
		{
			name:         "free text with validation fails",
			content:      "intro\n" + codeFileBlock("a.txt", "x"),
			validateCode: true,
			wantErr:      true,
		},
		{
			name:           "incomplete code block sets the incomplete flag",
			content:        fence + "\n// a.txt\nx\n",
			validateCode:   false,
			wantIncomplete: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var incomplete bool
			var err error
			output := captureStdout(t, func() {
				_, incomplete, err = parseBufferElements(tc.content, tc.validateCode)
			})

			if tc.wantErr {
				if err == nil {
					t.Fatal("parseBufferElements() error = nil, want error")
				}
			} else if err != nil {
				t.Fatalf("parseBufferElements() unexpected error: %v", err)
			}
			if incomplete != tc.wantIncomplete {
				t.Errorf("parseBufferElements() incomplete = %t, want %t", incomplete, tc.wantIncomplete)
			}
			if output != tc.wantOutput {
				t.Errorf("parseBufferElements() printed %q, want %q", output, tc.wantOutput)
			}
		})
	}
}