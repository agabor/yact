package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yact/config"
)

// The reply names files that the prompt does not link, so they must be rejected rather than written.
func TestActRejectsFilesNotLinkedInPrompt(t *testing.T) {
	prompt := readFixture(t, "test/hello-2/yact-prompt-0.md")
	reply := readFixture(t, "test/hello-2/.yact/temp/buffer.txt")
	planSystemPrompt := readFixture(t, "systemprompts/plan.txt")
	actSystemPrompt := readFixture(t, "systemprompts/act.txt")
	principles := readFixture(t, "systemprompts/include/principles.txt")

	useTempHome(t, planSystemPrompt, actSystemPrompt, principles)
	projectDir := enterTempProject(t)
	writeFile(t, filepath.Join(projectDir, "yact-prompt.md"), prompt)
	writeFile(t, filepath.Join(projectDir, "main.go"), "package main")

	client := newRecordingLLMClient(reply)
	err := runPhase("act", config.OutputNameApply, client)
	if err == nil {
		t.Fatalf("act phase succeeded with files the prompt does not link, want an error")
	}

	rejectedPaths := []string{"fibonacci/fibonacci.go", "fibonacci/fibonacci_test.go"}
	for _, path := range rejectedPaths {
		if !strings.Contains(err.Error(), path) {
			t.Fatalf("error %q does not name rejected file %s", err.Error(), path)
		}
	}

	if client.calls != 1 {
		t.Fatalf("act phase made %d LLM calls, want 1", client.calls)
	}

	for _, path := range rejectedPaths {
		if _, statErr := os.Stat(filepath.Join(projectDir, filepath.FromSlash(path))); !os.IsNotExist(statErr) {
			t.Fatalf("%s exists after the act phase rejected it, want it absent", path)
		}
	}
}
