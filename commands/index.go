package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const indexFileName = "index.txt"

type IndexNode struct {
	Name      string
	Type      string
	Functions []string
	Children  []*IndexNode
}

type indexLanguage struct {
	Extensions     []string
	ScanSignatures func(source string) []string
}

func lookupIndexLanguage(name string) (indexLanguage, error) {
	switch strings.ToLower(name) {
	case "go":
		return indexLanguage{Extensions: []string{".go"}, ScanSignatures: scanFuncSignatures}, nil
	case "php":
		return indexLanguage{Extensions: []string{".php"}, ScanSignatures: scanPHPSignatures}, nil
	case "js":
		return indexLanguage{Extensions: []string{".js", ".mjs", ".cjs"}, ScanSignatures: scanJSSignatures}, nil
	}
	return indexLanguage{}, fmt.Errorf("unsupported language %q (expected go, php or js)", name)
}

func hasLanguageExtension(name string, extensions []string) bool {
	for _, extension := range extensions {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}
	return false
}

func HandleIndexCommand(language string) error {
	indexLang, err := lookupIndexLanguage(language)
	if err != nil {
		return err
	}
	root, err := buildIndexNode(".", indexLang)
	if err != nil {
		return err
	}
	data := renderIndex(root)
	if err := os.WriteFile(indexFileName, []byte(data), 0644); err != nil {
		return err
	}
	fmt.Printf("Written: %s\n", indexFileName)
	return nil
}

func renderIndex(root *IndexNode) string {
	sections := collectFileSections(root, "", nil)
	return strings.Join(sections, "\n")
}

func collectFileSections(node *IndexNode, prefix string, sections []string) []string {
	for _, child := range node.Children {
		childPath := prefix + child.Name
		if child.Type == "dir" {
			sections = collectFileSections(child, childPath+"/", sections)
			continue
		}
		sections = append(sections, renderFileSection(childPath, child.Functions))
	}
	return sections
}

func renderFileSection(path string, functions []string) string {
	var builder strings.Builder
	builder.WriteString(path)
	builder.WriteString("\n")
	for _, function := range functions {
		builder.WriteString("  ")
		builder.WriteString(function)
		builder.WriteString("\n")
	}
	return builder.String()
}

func buildIndexNode(dir string, language indexLanguage) (*IndexNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	node := &IndexNode{
		Name: filepath.Base(dir),
		Type: "dir",
	}
	for _, entry := range entries {
		if shouldSkipIndexEntry(entry.Name()) {
			continue
		}
		childPath := filepath.Join(dir, entry.Name())
		child, err := buildChildIndexNode(childPath, entry, language)
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

func shouldSkipIndexEntry(name string) bool {
	return strings.HasPrefix(name, ".") ||
		name == "vendor" ||
		name == "node_modules" ||
		name == indexFileName
}

func buildChildIndexNode(path string, entry os.DirEntry, language indexLanguage) (*IndexNode, error) {
	if entry.IsDir() {
		return buildNonEmptyDirNode(path, language)
	}
	if hasLanguageExtension(entry.Name(), language.Extensions) {
		return buildSourceFileNode(path, entry.Name(), language)
	}
	return nil, nil
}

func buildNonEmptyDirNode(path string, language indexLanguage) (*IndexNode, error) {
	child, err := buildIndexNode(path, language)
	if err != nil {
		return nil, err
	}
	if len(child.Children) == 0 {
		return nil, nil
	}
	return child, nil
}

func buildSourceFileNode(path string, name string, language indexLanguage) (*IndexNode, error) {
	functions, err := extractFunctions(path, language)
	if err != nil {
		return nil, err
	}
	return &IndexNode{
		Name:      name,
		Type:      "file",
		Functions: functions,
	}, nil
}

func extractFunctions(path string, language indexLanguage) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return language.ScanSignatures(string(content)), nil
}

func scanFuncSignatures(source string) []string {
	return scanSignatures(source, isFuncDeclarationStart)
}

func scanSignatures(source string, isDeclarationStart func(source string, lineStart int) bool) []string {
	var signatures []string
	lineStart := 0
	for lineStart < len(source) {
		if !isDeclarationStart(source, lineStart) {
			lineStart = nextLineStart(source, lineStart)
			continue
		}
		signature, stop := signatureText(source, lineStart)
		declarationEnd := findDeclarationEnd(source, stop)
		signatures = append(signatures, signature)
		lineStart = nextLineStart(source, declarationEnd)
	}
	return signatures
}

func findDeclarationEnd(source string, stop int) int {
	if stop < len(source) && source[stop] == ';' {
		return stop + 1
	}
	return findFuncBodyEnd(source, stop)
}

func isFuncDeclarationStart(source string, lineStart int) bool {
	rest := source[lineStart:]
	if !strings.HasPrefix(rest, "func") && !strings.HasPrefix(rest, "type") {
		return false
	}
	after := rest[len("func"):]
	return strings.HasPrefix(after, " ") || strings.HasPrefix(after, "(")
}

func nextLineStart(source string, offset int) int {
	newline := strings.IndexByte(source[offset:], '\n')
	if newline < 0 {
		return len(source)
	}
	return offset + newline + 1
}

func findFuncBodyEnd(source string, open int) int {
	depth := 0
	for index := open; index < len(source); {
		switch {
		case strings.HasPrefix(source[index:], "//"):
			index = skipLineComment(source, index)
		case strings.HasPrefix(source[index:], "/*"):
			index = skipBlockComment(source, index)
		case isLiteralQuote(source[index]):
			index = skipLiteral(source, index)
		case source[index] == '{':
			depth++
			index++
		case source[index] == '}':
			depth--
			index++
			if depth == 0 {
				return index
			}
		default:
			index++
		}
	}
	return len(source)
}

func isLiteralQuote(char byte) bool {
	return char == '"' || char == '\'' || char == '`'
}

func skipLineComment(source string, start int) int {
	newline := strings.IndexByte(source[start:], '\n')
	if newline < 0 {
		return len(source)
	}
	return start + newline
}

func skipBlockComment(source string, start int) int {
	end := strings.Index(source[start+2:], "*/")
	if end < 0 {
		return len(source)
	}
	return start + 2 + end + 2
}

func skipLiteral(source string, start int) int {
	quote := source[start]
	for index := start + 1; index < len(source); index++ {
		if source[index] == '\\' && quote != '`' {
			index++
			continue
		}
		if source[index] == quote {
			return index + 1
		}
	}
	return len(source)
}

func signatureText(source string, start int) (string, int) {
	stop := strings.IndexAny(source[start:], "{;")
	if stop < 0 {
		return collapseWhitespace(source[start:]), len(source)
	}
	stop += start
	return collapseWhitespace(source[start:stop]), stop
}

func collapseWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}