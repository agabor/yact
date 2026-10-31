package commands

import (
	"testing"

	"yact/api"
)

func TestPrintBatchStatus(t *testing.T) {
	status := api.BatchStatus{
		ProcessingStatus: "in_progress",
		Processing:       3,
		Succeeded:        4,
		Errored:          5,
		Canceled:         6,
		Expired:          7,
	}

	got := captureStdout(t, func() {
		printBatchStatus(status)
	})

	want := "Batch status: in_progress\n" +
		"Processing: 3\n" +
		"Succeeded: 4\n" +
		"Errored: 5\n" +
		"Canceled: 6\n" +
		"Expired: 7\n"
	if got != want {
		t.Errorf("printBatchStatus() output = %q, want %q", got, want)
	}
}

func TestPrintBatchUsage(t *testing.T) {
	result := api.BatchResult{
		InputTokens:  1200,
		OutputTokens: 340,
		Cost:         0.012345,
	}

	got := captureStdout(t, func() {
		printBatchUsage(result)
	})

	want := "Input tokens: 1200\n" +
		"Output tokens: 340\n" +
		"Cost (50% batch discount applied): $0.0123\n"
	if got != want {
		t.Errorf("printBatchUsage() output = %q, want %q", got, want)
	}
}