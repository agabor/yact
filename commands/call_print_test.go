package commands

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"yact/api"
	"yact/config"
	"yact/logic"
)

func TestPrintWarnings(t *testing.T) {
	got := captureStdout(t, func() {
		printWarnings([]string{"first", "second"})
	})

	want := "Warning: first\nWarning: second\n"
	if got != want {
		t.Errorf("printWarnings() output = %q, want %q", got, want)
	}
}

func TestPrintSentFiles(t *testing.T) {
	requestContent := logic.RequestContent{
		Domain: "domain.txt",
		Output: "output.txt",
		Files:  []string{"a.go", "b.go"},
	}

	got := captureStdout(t, func() {
		printSentFiles(requestContent)
	})

	want := "Sending: domain.txt\nSending: output.txt\nSending: a.go\nSending: b.go\n"
	if got != want {
		t.Errorf("printSentFiles() output = %q, want %q", got, want)
	}
}

func TestPrintSentPrompt(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{name: "empty prompt prints nothing", prompt: "", want: ""},
		{name: "non-empty prompt prints header and prompt", prompt: "hello", want: "Sending prompt:\nhello\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				printSentPrompt(tc.prompt)
			})
			if got != tc.want {
				t.Errorf("printSentPrompt(%q) output = %q, want %q", tc.prompt, got, tc.want)
			}
		})
	}
}

func TestPrintCallResult(t *testing.T) {
	result := api.CallResult{
		Duration:     1500 * time.Millisecond,
		FileCount:    2,
		InputTokens:  100,
		OutputTokens: 50,
		Cost:         0.001234,
	}
	summary := "Claude API call took 1.50 seconds\n" +
		"Files sent: 2\n" +
		"Token usage - Input: 100, Output: 50\n" +
		"Cost: $0.001234\n"
	truncationWarning := fmt.Sprintf("WARNING: Maximum output tokens (%d) reached", config.MaxTokens)

	t.Run("prints token, cost and file lines", func(t *testing.T) {
		got := captureStdout(t, func() {
			printCallResult(result)
		})
		if got != summary {
			t.Errorf("printCallResult() output = %q, want %q", got, summary)
		}
	})

	t.Run("prints truncation warning only when truncated", func(t *testing.T) {
		truncated := result
		truncated.Truncated = true
		got := captureStdout(t, func() {
			printCallResult(truncated)
		})
		if !strings.Contains(got, truncationWarning) {
			t.Errorf("printCallResult() output = %q, want it to contain %q", got, truncationWarning)
		}
	})
}

func TestPrintApplyOutcome(t *testing.T) {
	path := logic.ProjectPath("src/a.go")
	tests := []struct {
		name    string
		outcome logic.ApplyOutcome
		want    string
	}{
		{name: "written", outcome: logic.OutcomeWritten, want: "Written: src/a.go\n"},
		{name: "deleted", outcome: logic.OutcomeDeleted, want: "Deleted: src/a.go\n"},
		{name: "nothing to delete", outcome: logic.OutcomeNothingToDelete, want: "Not found (nothing to delete): src/a.go\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				printApplyOutcome(path, tc.outcome)
			})
			if got != tc.want {
				t.Errorf("printApplyOutcome() output = %q, want %q", got, tc.want)
			}
		})
	}
}