package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"yact/logic"
)

type IndexNode struct {
	Name      string
	Type      string
	Functions []IndexSignature
	Children  []*IndexNode
}

type IndexSignature struct {
	Text  string
	Depth int
}

type signatureRules struct {
	IsDeclarationStart  func(source string, lineStart int) bool
	IsMemberStart       func(source string, lineStart int) bool
	IsContainerStart    func(source string, lineStart int) bool
	PublicOnly          bool
	IsPublicDeclaration func(source string, lineStart int) bool
	IsPublicMember      func(source string, lineStart int) bool
	EndsAtLineBreak     bool
}

type indexLanguage struct {
	Extensions []string
	Rules      signatureRules
}

var goSignatureRules = signatureRules{
	IsDeclarationStart:  isFuncDeclarationStart,
	IsMemberStart:       isFuncDeclarationStart,
	IsPublicDeclaration: isGoPublicDeclaration,
	IsPublicMember:      isGoPublicDeclaration,
}

var indexLanguages = []indexLanguage{
	{Extensions: []string{".go"}, Rules: goSignatureRules},
	{Extensions: []string{".php"}, Rules: phpSignatureRules},
	{Extensions: []string{".js", ".mjs", ".cjs"}, Rules: jsSignatureRules},
	{Extensions: []string{".cs"}, Rules: csharpSignatureRules},
	{Extensions: []string{".swift"}, Rules: swiftSignatureRules},
	{Extensions: []string{".kt", ".kts"}, Rules: kotlinSignatureRules},
}

const indexHeaderIntroduction = `This is an index of the source code in the project.
Each section starts with a file path relative to the project root, followed by
the signatures of the declarations (functions, methods, types, classes) found in that file.
Function bodies are omitted. Indented signatures are members of the declaration above them.
Use this index to understand the structure of the project and to locate code.`

const publicIndexScope = "Only the public interface is listed: private declarations are omitted."

const privateIndexScope = "All declarations are listed, including private ones."

func languagesForVisibility(visibility string) ([]indexLanguage, error) {
	if visibility != "public" && visibility != "private" {
		return nil, fmt.Errorf("invalid visibility %q, expected private or public", visibility)
	}
	languages := make([]indexLanguage, len(indexLanguages))
	copy(languages, indexLanguages)
	for position := range languages {
		languages[position].Rules.PublicOnly = visibility == "public"
	}
	return languages, nil
}

func findLanguageForFile(name string, languages []indexLanguage) (indexLanguage, bool) {
	for _, language := range languages {
		if hasLanguageExtension(name, language.Extensions) {
			return language, true
		}
	}
	return indexLanguage{}, false
}

func hasLanguageExtension(name string, extensions []string) bool {
	for _, extension := range extensions {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}
	return false
}

func WriteIndex(visibility string) error {
	languages, err := languagesForVisibility(visibility)
	if err != nil {
		return err
	}
	root, err := buildIndexNode(".", languages)
	if err != nil {
		return err
	}
	data := renderIndex(root, visibility)
	indexPath := logic.IndexFilePath()
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(indexPath, []byte(data), 0644)
}

func indexLevelToVisibility(level int) (string, error) {
	switch level {
	case 1:
		return "public", nil
	case 2:
		return "private", nil
	}
	return "", fmt.Errorf("invalid index level %d, expected 1 or 2", level)
}

func renderIndex(root *IndexNode, visibility string) string {
	sections := collectFileSections(root, "", nil)
	return renderIndexHeader(visibility) + "\n" + strings.Join(sections, "\n")
}

func renderIndexHeader(visibility string) string {
	return indexHeaderIntroduction + "\n" + indexScopeDescription(visibility) + "\n"
}

func indexScopeDescription(visibility string) string {
	if visibility == "public" {
		return publicIndexScope
	}
	return privateIndexScope
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

func renderFileSection(path string, signatures []IndexSignature) string {
	var builder strings.Builder
	builder.WriteString(path)
	builder.WriteString("\n")
	for _, signature := range signatures {
		builder.WriteString("  ")
		builder.WriteString(strings.Repeat("  ", signature.Depth))
		builder.WriteString(signature.Text)
		builder.WriteString("\n")
	}
	return builder.String()
}

func buildIndexNode(dir string, languages []indexLanguage) (*IndexNode, error) {
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
		child, err := buildChildIndexNode(childPath, entry, languages)
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
		name == "node_modules"
}

func buildChildIndexNode(path string, entry os.DirEntry, languages []indexLanguage) (*IndexNode, error) {
	if entry.IsDir() {
		return buildNonEmptyDirNode(path, languages)
	}
	language, found := findLanguageForFile(entry.Name(), languages)
	if !found {
		return nil, nil
	}
	return buildSourceFileNode(path, entry.Name(), language)
}

func buildNonEmptyDirNode(path string, languages []indexLanguage) (*IndexNode, error) {
	child, err := buildIndexNode(path, languages)
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

func extractFunctions(path string, language indexLanguage) ([]IndexSignature, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return scanSignatures(string(content), language.Rules), nil
}

func scanSignatures(source string, rules signatureRules) []IndexSignature {
	return scanSignatureBlock(source, 0, len(source), 0, rules, false)
}

func scanSignatureBlock(source string, start int, end int, depth int, rules signatureRules, inContainer bool) []IndexSignature {
	isStart := selectDeclarationMatcher(rules, inContainer)
	var signatures []IndexSignature
	lineStart := start
	for lineStart < end {
		if !isStart(source, lineStart) {
			lineStart = nextLineStart(source, lineStart)
			continue
		}
		signature, stop := signatureText(source, lineStart, rules.EndsAtLineBreak)
		declarationEnd := findDeclarationEnd(source, stop)
		if !isVisibleDeclaration(source, lineStart, rules, inContainer) {
			lineStart = nextLineStart(source, declarationEnd)
			continue
		}
		signatures = append(signatures, IndexSignature{Text: signature, Depth: depth})
		if isContainerDeclaration(source, lineStart, stop, rules) {
			signatures = append(signatures, scanSignatureBlock(source, stop+1, declarationEnd-1, depth+1, rules, true)...)
		}
		lineStart = nextLineStart(source, declarationEnd)
	}
	return signatures
}

func selectDeclarationMatcher(rules signatureRules, inContainer bool) func(source string, lineStart int) bool {
	if inContainer && rules.IsMemberStart != nil {
		return rules.IsMemberStart
	}
	return rules.IsDeclarationStart
}

func selectPublicMatcher(rules signatureRules, inContainer bool) func(source string, lineStart int) bool {
	if inContainer {
		return rules.IsPublicMember
	}
	return rules.IsPublicDeclaration
}

func isVisibleDeclaration(source string, lineStart int, rules signatureRules, inContainer bool) bool {
	if !rules.PublicOnly {
		return true
	}
	isPublic := selectPublicMatcher(rules, inContainer)
	if isPublic == nil {
		return true
	}
	return isPublic(source, lineStart)
}

func isContainerDeclaration(source string, lineStart int, stop int, rules signatureRules) bool {
	if rules.IsContainerStart == nil || !rules.IsContainerStart(source, lineStart) {
		return false
	}
	return stop < len(source) && source[stop] == '{'
}

func findDeclarationEnd(source string, stop int) int {
	if stop < len(source) && source[stop] == ';' {
		return stop + 1
	}
	if stop < len(source) && source[stop] == '\n' {
		return stop
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

func isGoPublicDeclaration(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart+len("func"))
	if strings.HasPrefix(source[lineStart:], "func") && position < len(source) && source[position] == '(' {
		position = skipIndentation(source, skipGoReceiver(source, position))
	}
	if position >= len(source) {
		return false
	}
	first, _ := utf8.DecodeRuneInString(source[position:])
	return unicode.IsUpper(first)
}

func skipGoReceiver(source string, position int) int {
	closing := strings.IndexByte(source[position:], ')')
	if closing < 0 {
		return len(source)
	}
	return position + closing + 1
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

func signatureText(source string, start int, endsAtLineBreak bool) (string, int) {
	terminators := "{;"
	if endsAtLineBreak {
		terminators = "{;\n"
	}
	stop := strings.IndexAny(source[start:], terminators)
	if stop < 0 {
		return collapseWhitespace(source[start:]), len(source)
	}
	stop += start
	return collapseWhitespace(source[start:stop]), stop
}

func collapseWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
