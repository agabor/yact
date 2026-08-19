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

type CodeFile struct {
	Path    string
	Content string
}

type codeFileJSON struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func NewCodeFile(path, content string) CodeFile {
	return CodeFile{
		Path:    strings.TrimPrefix(strings.TrimSpace(path), "./"),
		Content: content,
	}
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
			extracted := strings.TrimSpace(matches[1])
			if strings.Contains(extracted, "!") {
				continue
			}
			return regexp.MustCompile(`\s*\*+/$`).ReplaceAllString(extracted, "")
		}
	}
	return ""
}

func detectFilePath(lines []string) (string, int) {
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
			return extractedPath, lineIndex
		}
	}

	return "", -1
}

func isValidCodeFile(lines []string) bool {
	path, _ := detectFilePath(lines)
	return path != ""
}

func linesToCodeFile(lines []string) CodeFile {
	filePath, lineIndex := detectFilePath(lines)

	if filePath != "" {
		lines = append(lines[:lineIndex], lines[lineIndex+1:]...)
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

	return CodeFile{
		Path:    strings.TrimPrefix(strings.TrimSpace(filePath), "./"),
		Content: joinLines(lines),
	}
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func ReadAsFile(filePath string) (CodeFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return CodeFile{}, err
	}

	return CodeFile{
		Path:    filePath,
		Content: string(content),
	}, nil
}

func (cf *CodeFile) Write() error {
	filePath := cf.Path
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(cf.Content), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}

	fmt.Printf("Written: %s\n", filePath)
	return nil
}

func ParseCodeFiles(response string) []interface{} {
	lines := strings.Split(response, "\n")
	var result = make([]interface{}, 0)
	var lineBuffer = make([]string, 0)
	var textBuffer = make([]string, 0)
	inBlock := false

	flushText := func() {
		if len(textBuffer) > 0 {
			result = append(result, strings.Join(textBuffer, "\n"))
			textBuffer = make([]string, 0)
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), BlockDelimiterMin) {
			if inBlock {
				if len(lineBuffer) > 0 {
					if isValidCodeFile(lineBuffer) {
						result = append(result, linesToCodeFile(lineBuffer))
					} else {
						result = append(result, joinLines(lineBuffer))
					}
				}
				inBlock = false
				lineBuffer = make([]string, 0)
			} else {
				flushText()
				inBlock = true
			}
		} else if inBlock {
			lineBuffer = append(lineBuffer, line)
		} else {
			textBuffer = append(textBuffer, line)
		}
	}

	flushText()

	if inBlock && len(lineBuffer) > 0 {
		if isValidCodeFile(lineBuffer) {
			result = append(result, linesToCodeFile(lineBuffer))
		} else {
			result = append(result, joinLines(lineBuffer))
		}
		fmt.Println("Warning: Incomplete code block.")
	}

	return result
}