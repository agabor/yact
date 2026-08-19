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

	transaction := logic.Transaction{
		Request: []string{snippet, prompt},
	}
	response, err := callClaudeAPI(transaction, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	elements := logic.ParseCodeFiles(response)

	var firstCodeFile *logic.CodeFile
	for _, element := range elements {
		if codeFile, ok := element.(logic.CodeFile); ok {
			firstCodeFile = &codeFile
			break
		}
	}

	if firstCodeFile == nil {
		return fmt.Errorf("no code blocks found in Claude's response")
	}

	indent := getMinIndentation(snippet)
	replacement := applyIndentation(firstCodeFile.Content, indent)

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

func getMinIndentation(code string) int {
	lines := strings.Split(code, "\n")
	minIndent := -1

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		leadingSpaces := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || leadingSpaces < minIndent {
			minIndent = leadingSpaces
		}
	}

	if minIndent == -1 {
		return 0
	}

	return minIndent
}

func applyIndentation(code string, targetIndent int) string {
	lines := strings.Split(code, "\n")
	currentIndent := getMinIndentation(code)
	indentDiff := targetIndent - currentIndent

	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
			continue
		}

		if indentDiff >= 0 {
			result = append(result, strings.Repeat(" ", indentDiff)+line)
		} else {
			result = append(result, strings.TrimLeft(line, " \t"))
			result[len(result)-1] = strings.Repeat(" ", targetIndent) + strings.TrimLeft(line, " \t")
		}
	}

	return strings.Join(result, "\n")
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