package commands

import (
	"testing"

	"yact/logic"
)

const fence = "```"

func codeFileBlock(path string, body string) string {
	return fence + "\n// " + path + "\n" + body + "\n" + fence + "\n"
}

func TestProcessResponsePrintsTextOnce(t *testing.T) {
	elements := []logic.ParsedElement{
		logic.NewTextElement("intro"),
	}

	var err error
	got := captureStdout(t, func() {
		err = processResponse(elements)
	})

	if err != nil {
		t.Fatalf("processResponse() unexpected error: %v", err)
	}
	want := "intro\n"
	if got != want {
		t.Errorf("processResponse() printed %q, want %q", got, want)
	}
}
