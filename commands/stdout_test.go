package commands

import (
	"io"
	"os"
	"testing"
)

// Tests that use captureStdout must not call t.Parallel, because os.Stdout is global.
func captureStdout(t *testing.T, action func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}

	originalStdout := os.Stdout
	defer func() {
		os.Stdout = originalStdout
	}()
	os.Stdout = writer

	captured := make(chan string)
	go func() {
		data, _ := io.ReadAll(reader)
		captured <- string(data)
	}()

	action()

	writer.Close()
	output := <-captured
	reader.Close()
	return output
}