package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"yact/logic"
)

func TestReadAsCodeBlock(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := logic.ReadAsCodeBlock(testFile)
	if err != nil {
		t.Fatalf("ReadAsCodeBlock failed: %v", err)
	}

	if !strings.Contains(result, logic.BlockDelimiter) {
		t.Error("result does not contain block delimiter")
	}

	if !strings.Contains(result, testFile) {
		t.Error("result does not contain file path")
	}

	if !strings.Contains(result, testContent) {
		t.Error("result does not contain file content")
	}
}

func TestReadAsCodeBlockFileNotFound(t *testing.T) {
	_, err := logic.ReadAsCodeBlock("/nonexistent/file/path.txt")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestReadAsCodeBlockEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := logic.ReadAsCodeBlock(testFile)
	if err != nil {
		t.Fatalf("ReadAsCodeBlock failed: %v", err)
	}

	if !strings.Contains(result, logic.BlockDelimiter) {
		t.Error("result does not contain block delimiter")
	}

	if !strings.Contains(result, testFile) {
		t.Error("result does not contain file path")
	}
}
