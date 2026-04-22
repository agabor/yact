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
	path    string
	Content string
}

type codeFileJSON struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func NewCodeFile(path, content string) CodeFile {
	return CodeFile{
		path:    strings.TrimPrefix(strings.TrimSpace(path), "./"),
		Content: content,
	}
}

func (f CodeFile) Path() string {
	return f.path
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

func linesToCodeFile(lines []string) CodeFile {
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

	return NewCodeFile(filePath, joinLines(lines))
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func ReadAsFile(filePath string) (CodeFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return CodeFile{}, err
	}

	return NewCodeFile(filePath, string(content)), nil
}

func (cf *CodeFile) Write() error {
	filePath := cf.Path()
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

func ParseCodeFiles(response string) ([]CodeFile, []string) {
	lines := strings.Split(response, "\n")
	var codeFiles = make([]CodeFile, 0)
	var lineBuffer = make([]string, 0)
	var text = make([]string, 0)
	inBlock := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), BlockDelimiterMin) {
			if inBlock {
				if len(lineBuffer) > 0 {
					codeFiles = append(codeFiles, linesToCodeFile(lineBuffer))
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
		codeFile := linesToCodeFile(lineBuffer)
		codeFiles = append(codeFiles, codeFile)
		fmt.Println("Warning: Incomplete code block.")
	}

	return codeFiles, text
}
