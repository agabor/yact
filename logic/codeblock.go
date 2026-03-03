package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var unknownFileCounter = 0

const BlockDelimiterMin = "``" + "`"
const BlockDelimiter = BlockDelimiterMin + "`"

type CodeBlock struct {
	Path    string
	Content string
}

func extractFilenameFromComment(line string) string {
	patterns := []string{
		`^\s*//\s*(.+?)(?:\s*//.*)?$`,
		`^\s*#\s*(.+?)(?:\s*#.*)?$`,
		`^\s*#\s*//\s*(.+?)(?:\s*#.*)?$`,
		`^\s*/\*\s*(.+?)\s*\*/$`,
		`^\s*--\s*(.+?)(?:\s*--.*)?$`,
		`^\s*<!--\s*(.+?)\s*-->$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			filename := strings.TrimSpace(matches[1])
			if strings.Contains(filename, "!") {
				continue
			}
			filename = regexp.MustCompile(`\s*\*+/$`).ReplaceAllString(filename, "")
			return filename
		}
	}
	return ""
}

func linesToCodeBlock(lines []string) CodeBlock {
	filePath := ""
	lineIndex := 0

	for lineIndex < len(lines) && strings.TrimSpace(lines[lineIndex]) == "" {
		lineIndex++
	}

	if lineIndex < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[lineIndex]), "#!") {
		lineIndex++
	}

	if lineIndex < len(lines) {
		extractedPath := extractFilenameFromComment(lines[lineIndex])
		if extractedPath != "" {
			filePath = extractedPath
			lines = append(lines[:lineIndex], lines[lineIndex+1:]...)
		}
	}

	if filePath == "" {
		unknownFileCounter += 1
		filePath = "unknown" + strconv.Itoa(unknownFileCounter)
	}

	if strings.HasPrefix(filePath, "/") {
		relPath := filePath[1:]
		if _, err := os.Stat(relPath); err == nil {
			filePath = relPath
		} else {
			if _, err := os.Stat(filePath); err == nil {

			} else {
				filePath = relPath
			}
		}
	}

	return CodeBlock{Path: filePath, Content: joinLines(lines)}
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func AsCodeBlock(path string, content string) string {
	return joinLines([]string{BlockDelimiter, "//" + path, content, BlockDelimiter})
}

func ReadAsCodeBlock(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return AsCodeBlock(filePath, string(content)), nil
}

func (cb *CodeBlock) Write(safe bool) error {
	filePath := cb.Path
	if safe {
		filePath += ".new"
	}
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(cb.Content), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}

	fmt.Printf("Written: %s\n", filePath)
	return nil
}
func ParseCodeBlocks(response string) ([]CodeBlock, []string) {
	lines := strings.Split(response, "\n")
	var codeBlocks = make([]CodeBlock, 0)
	var lineBuffer = make([]string, 0)
	var text = make([]string, 0)
	inBlock := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), BlockDelimiterMin) {
			if inBlock {
				if len(lineBuffer) > 0 {
					codeBlocks = append(codeBlocks, linesToCodeBlock(lineBuffer))
				}
				inBlock = false
				lineBuffer = make([]string, 0)
			} else {
				inBlock = true
			}
		} else if inBlock {
			lineBuffer = append(lineBuffer, line)
		} else {
			text = append(text, line)
		}
	}

	if inBlock && len(lineBuffer) > 0 {
		codeBlocks = append(codeBlocks, linesToCodeBlock(lineBuffer))
	}

	return codeBlocks, text
}
