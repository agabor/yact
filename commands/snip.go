package commands

import (
	"fmt"
	"os"
	"strings"

	"yact/config"
	"yact/logic"
)

func HandleSnipCommand(inputFile string, startLine int, endLine int, prompt string, think bool, cfg *config.Config, systemPrompt string) error {
	if startLine <= 0 || endLine <= 0 {
		return fmt.Errorf("line numbers must be positive")
	}

	if startLine > endLine {
		return fmt.Errorf("start line cannot be greater than end line")
	}

	snippet, err := readLinesFromFile(inputFile, startLine, endLine)
	if err != nil {
		return err
	}

	transactions := []logic.Transaction{
		{
			Request:  []string{snippet, prompt},
			Response: "",
		},
	}

	transactions, err = callClaudeAPI("", transactions, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	response := transactions[len(transactions)-1].Response
	codeblocks, _ := logic.ParseCodeBlocks(response)

	if len(codeblocks) == 0 {
		return fmt.Errorf("no code blocks found in Claude's response")
	}

	replacement := codeblocks[0].Content

	err = replaceLineRange(inputFile, startLine, endLine, replacement)
	if err != nil {
		return err
	}

	fmt.Println("Done!")
	return nil
}

func readLinesFromFile(filename string, startLine int, endLine int) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	if startLine > len(lines) || endLine > len(lines) {
		return "", fmt.Errorf("line range exceeds file length (%d lines)", len(lines))
	}

	selectedLines := lines[startLine-1 : endLine]
	snippet := strings.Join(selectedLines, "\n")

	return snippet, nil
}

func replaceLineRange(filename string, startLine int, endLine int, replacement string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	if startLine > len(lines) || endLine > len(lines) {
		return fmt.Errorf("line range exceeds file length (%d lines)", len(lines))
	}

	replacementLines := strings.Split(replacement, "\n")

	newLines := append(lines[:startLine-1], append(replacementLines, lines[endLine:]...)...)
	newContent := strings.Join(newLines, "\n")

	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}