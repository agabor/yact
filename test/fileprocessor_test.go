package test

import (
	"strings"
	"testing"
	"yact/logic"
)

func TestParseCodeBlocksBasic(t *testing.T) {
	input := "````\n// src/main.go\npackage main\n````"
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 1 {
		t.Errorf("expected 1 code block, got %d", len(blocks))
	}

	if !strings.Contains(blocks[0].Path, "main.go") {
		t.Error("expected file path to contain main.go")
	}

	if !strings.Contains(blocks[0].Content, "package main") {
		t.Error("expected content to contain 'package main'")
	}
}

func TestParseCodeBlocksMultiple(t *testing.T) {
	input := "````\n// file1.go\ncontent1\n````\n\n````\n// file2.go\ncontent2\n````"
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 2 {
		t.Errorf("expected 2 code blocks, got %d", len(blocks))
	}

	if !strings.Contains(blocks[0].Path, "file1.go") {
		t.Error("expected first block file path to contain file1.go")
	}

	if !strings.Contains(blocks[1].Path, "file2.go") {
		t.Error("expected second block file path to contain file2.go")
	}
}

func TestParseCodeBlocksEmpty(t *testing.T) {
	input := ""
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 0 {
		t.Errorf("expected 0 code blocks for empty input, got %d", len(blocks))
	}
}

func TestParseCodeBlocksNoBlocks(t *testing.T) {
	input := "just some text\nwithout any code blocks"
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 0 {
		t.Errorf("expected 0 code blocks, got %d", len(blocks))
	}
}

func TestParseCodeBlocksIncompleteBlock(t *testing.T) {
	input := "````\n// file.go\ncontent\n"
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 1 {
		t.Errorf("expected 1 code block for incomplete delimiter, got %d", len(blocks))
	}
}

func TestParseCodeBlocksWithExtraText(t *testing.T) {
	input := "some text before\n````\n// file.go\ncode\n````\nsome text after"
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 1 {
		t.Errorf("expected 1 code block, got %d", len(blocks))
	}

	if !strings.Contains(blocks[0].Content, "code") {
		t.Error("expected content to contain 'code'")
	}
}

func TestParseCodeBlocksMultilineContent(t *testing.T) {
	input := "````\n// src/test.go\nfunc main() {\n  fmt.Println(\"hello\")\n}\n````"
	blocks, _, _ := logic.ParseCodeBlocks(input)

	if len(blocks) != 1 {
		t.Errorf("expected 1 code block, got %d", len(blocks))
	}

	content := blocks[0].Content
	if !strings.Contains(content, "func main()") {
		t.Error("expected content to contain function declaration")
	}

	if !strings.Contains(content, "fmt.Println") {
		t.Error("expected content to contain print statement")
	}
}