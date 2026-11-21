package logic

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"yact/config"

	"yact/projectpath"
)

const BlockDelimiterMin = "``" + "`"
const BlockDelimiter = BlockDelimiterMin + "`"

type CodeFile struct {
	Path    projectpath.Path
	Content string
}

type ParsedElementKind int

const (
	ElementText ParsedElementKind = iota
	ElementCodeFile
	ElementRejectedFile
)

type ParsedElement struct {
	Kind     ParsedElementKind
	CodeFile CodeFile
	Text     string
}

type ApplyOutcome int

const (
	OutcomeWritten ApplyOutcome = iota
	OutcomeDeleted
	OutcomeNothingToDelete
)

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

// A block is a code file only when its extension is listed; otherwise a code
// sample that starts with a filename comment would be written to disk.
func isValidCodeFile(lines []string, extensions []string) bool {
	p, _ := detectFilePath(lines)
	return p != "" && FileHasExtension(p, extensions)
}

func linesToCodeFile(lines []string) (CodeFile, error) {
	filePath, lineIndex := detectFilePath(lines)

	if filePath != "" {
		lines = append(lines[:lineIndex], lines[lineIndex+1:]...)
	}

	projectPath, err := projectpath.Parse(filePath)
	if err != nil {
		return CodeFile{}, err
	}

	return CodeFile{
		Path:    projectPath,
		Content: joinLines(lines),
	}, nil
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func ReadAsFile(filePath projectpath.Path) (CodeFile, error) {
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

func (cf *CodeFile) writeContent() error {
	filePath := string(cf.Path)
	dir := path.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(cf.Content), 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}

	return nil
}

func (cf *CodeFile) removeFile() (ApplyOutcome, error) {
	if cf.Path == "" {
		return OutcomeNothingToDelete, fmt.Errorf("cannot delete: empty path")
	}

	err := os.Remove(string(cf.Path))
	if err != nil {
		if os.IsNotExist(err) {
			return OutcomeNothingToDelete, nil
		}
		return OutcomeNothingToDelete, fmt.Errorf("error deleting file %s: %w", cf.Path, err)
	}

	return OutcomeDeleted, nil
}

// Empty content is how the LLM asks for a file to be removed, so it is not written.
func (cf *CodeFile) Apply() (ApplyOutcome, error) {
	if cf.IsEmpty() {
		return cf.removeFile()
	}
	if err := cf.writeContent(); err != nil {
		return OutcomeWritten, err
	}
	return OutcomeWritten, nil
}

// A response may write only the files the prompt linked, so a code block for any other path is rejected rather than written.
func processCodeBlock(lineBuffer []string, extensions []string, linked LinkedFiles) *ParsedElement {
	if len(lineBuffer) == 0 {
		return nil
	}
	if isValidCodeFile(lineBuffer, extensions) {
		codeFile, err := linesToCodeFile(lineBuffer)
		if err != nil {
			element := NewRejectedFileElement(err)
			return &element
		}
		if !linked.Allows(codeFile.Path) {
			element := NewRejectedFileElement(fmt.Errorf("file %s is not linked in the prompt", codeFile.Path))
			return &element
		}
		element := NewCodeFileElement(codeFile)
		return &element
	}
	element := NewTextElement(joinLines(lineBuffer))
	return &element
}

func ParseCodeFilesDetailed(response string, linked LinkedFiles) ([]ParsedElement, error) {
	extensions, err := config.LoadExtensions()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(response, "\n")
	var result = make([]ParsedElement, 0)
	var lineBuffer = make([]string, 0)

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), BlockDelimiterMin) {
			if block := processCodeBlock(lineBuffer, extensions, linked); block != nil {
				result = append(result, *block)
			}
			lineBuffer = make([]string, 0)
		} else {
			lineBuffer = append(lineBuffer, line)
		}
	}

	if block := processCodeBlock(lineBuffer, extensions, linked); block != nil {
		result = append(result, *block)
		lastIdx := len(result) - 1
		lastElement := result[lastIdx]
		if lastElement.Kind == ElementCodeFile {
			result[lastIdx] = NewRejectedFileElement(fmt.Errorf("file at %s is truncated", lastElement.CodeFile.Path))
		}
	}

	return result, nil
}
