package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const BlockDelimiterMin = "``" + "`"
const BlockDelimiter = BlockDelimiterMin + "`"

type CodeFile struct {
	Path    ProjectPath
	Content string
}

type ParsedElementKind int

const (
	ElementText ParsedElementKind = iota
	ElementCodeFile
	ElementRejectedFile
)

type ParsedElement struct {
	Kind       ParsedElementKind
	CodeFile   CodeFile
	Text       string
}

func NewCodeFileElement(cf CodeFile) ParsedElement {
	return ParsedElement{
		Kind:     ElementCodeFile,
		CodeFile: cf,
	}
}

func NewTextElement(text string) ParsedElement {
	return ParsedElement{
		Kind: ElementText,
		Text: text,
	}
}

func NewRejectedFileElement(reason error) ParsedElement {
	return ParsedElement{
		Kind: ElementRejectedFile,
		Text: reason.Error(),
	}
}

func extractFilenameFromComment(line string) string {
	patterns := []string{
		`^\s*//\s*(.+?)(?:\s*//.*)?$`,
		`^\s*#\s*(.+?)(?:\s*#.*)?$`,
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

func linesToCodeFile(lines []string) (CodeFile, error) {
	filePath, lineIndex := detectFilePath(lines)

	if filePath != "" {
		lines = append(lines[:lineIndex], lines[lineIndex+1:]...)
	}

	path, err := ParseProjectPath(filePath)
	if err != nil {
		return CodeFile{}, err
	}

	return CodeFile{
		Path:    path,
		Content: joinLines(lines),
	}, nil
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func ReadAsFile(filePath ProjectPath) (CodeFile, error) {
	content, err := os.ReadFile(string(filePath))
	if err != nil {
		return CodeFile{}, err
	}

	return CodeFile{
		Path:    filePath,
		Content: string(content),
	}, nil
}

func (cf *CodeFile) IsEmpty() bool {
	return strings.TrimSpace(cf.Content) == ""
}

func (cf *CodeFile) Write() error {
	filePath := string(cf.Path)
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

func (cf *CodeFile) Delete() error {
	if cf.Path == "" {
		return fmt.Errorf("cannot delete: empty path")
	}

	err := os.Remove(string(cf.Path))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Not found (nothing to delete): %s\n", cf.Path)
			return nil
		}
		return fmt.Errorf("error deleting file %s: %w", cf.Path, err)
	}

	fmt.Printf("Deleted: %s\n", cf.Path)
	return nil
}

func (cf *CodeFile) Apply() error {
	if cf.IsEmpty() {
		return cf.Delete()
	}
	return cf.Write()
}

func processCodeBlock(lineBuffer []string) *ParsedElement {
	if len(lineBuffer) == 0 {
		return nil
	}
	if isValidCodeFile(lineBuffer) {
		codeFile, err := linesToCodeFile(lineBuffer)
		if err != nil {
			element := NewRejectedFileElement(err)
			return &element
		}
		element := NewCodeFileElement(codeFile)
		return &element
	}
	element := NewTextElement(joinLines(lineBuffer))
	return &element
}

func ParseCodeFilesDetailed(response string) ([]ParsedElement, bool) {
	lines := strings.Split(response, "\n")
	var result = make([]ParsedElement, 0)
	var lineBuffer = make([]string, 0)
	var textBuffer = make([]string, 0)
	inBlock := false

	flushText := func() {
		if len(textBuffer) > 0 {
			result = append(result, NewTextElement(strings.Join(textBuffer, "\n")))
			textBuffer = make([]string, 0)
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), BlockDelimiterMin) {
			if inBlock {
				if block := processCodeBlock(lineBuffer); block != nil {
					result = append(result, *block)
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

	incomplete := false
	if inBlock {
		if block := processCodeBlock(lineBuffer); block != nil {
			result = append(result, *block)
			incomplete = true
		}
	}

	return result, incomplete
}

func ParseCodeFiles(response string) []ParsedElement {
	result, incomplete := ParseCodeFilesDetailed(response)
	if incomplete {
		fmt.Println("Warning: Incomplete code block.")
	}
	return result
}