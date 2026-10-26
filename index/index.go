package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	IsFunction        func(label string) bool
	IsPublic          func(label string) bool
	CleanLabel        func(label string) string
	SplitsAtLineBreak bool
	PublicOnly        bool
}

type indexLanguage struct {
	Extensions []string
	Rules      signatureRules
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
	files, err := logic.EnumerateFiles(".", shouldSkipIndexDir, func(name string) bool {
		_, found := findLanguageForFile(name, languages)
		return found
	})
	if err != nil {
		return err
	}
	root, err := buildIndexTree(files, languages)
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

func shouldSkipIndexDir(name string) bool {
	return name == "vendor" || name == "node_modules"
}

func buildIndexTree(paths []string, languages []indexLanguage) (*IndexNode, error) {
	root := &IndexNode{
		Name: ".",
		Type: "dir",
	}
	for _, path := range paths {
		name := filepath.Base(path)
		language, found := findLanguageForFile(name, languages)
		if !found {
			continue
		}
		fileNode, err := buildSourceFileNode(path, name, language)
		if err != nil {
			return nil, err
		}
		directory := ensureDirectoryNode(root, filepath.Dir(path))
		directory.Children = append(directory.Children, fileNode)
	}
	return root, nil
}

func ensureDirectoryNode(root *IndexNode, dirPath string) *IndexNode {
	if dirPath == "." {
		return root
	}
	current := root
	for _, name := range strings.Split(filepath.ToSlash(dirPath), "/") {
		current = findOrAddChildDirectory(current, name)
	}
	return current
}

func findOrAddChildDirectory(parent *IndexNode, name string) *IndexNode {
	for _, child := range parent.Children {
		if child.Type == "dir" && child.Name == name {
			return child
		}
	}
	child := &IndexNode{
		Name: name,
		Type: "dir",
	}
	parent.Children = append(parent.Children, child)
	return child
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
	root := parseBraceTree(source, rules.SplitsAtLineBreak)
	return collectSignatures(root, 0, rules)
}

func collectSignatures(node *braceNode, depth int, rules signatureRules) []IndexSignature {
	var signatures []IndexSignature
	for _, child := range node.Children {
		label := cleanLabel(child.Label, rules)
		if label == "" {
			signatures = append(signatures, collectSignatures(child, depth, rules)...)
			continue
		}
		if rules.IsFunction(label) {
			if isVisibleFunction(label, rules) {
				signatures = append(signatures, IndexSignature{Text: label, Depth: depth})
			}
			continue
		}
		nested := collectSignatures(child, depth+1, rules)
		if len(nested) > 0 {
			signatures = append(signatures, IndexSignature{Text: label, Depth: depth})
			signatures = append(signatures, nested...)
		}
	}
	return signatures
}

func cleanLabel(label string, rules signatureRules) string {
	if rules.CleanLabel == nil {
		return label
	}
	return rules.CleanLabel(label)
}

func isVisibleFunction(label string, rules signatureRules) bool {
	if !rules.PublicOnly || rules.IsPublic == nil {
		return true
	}
	return rules.IsPublic(label)
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

func isLiteralQuote(char byte) bool {
	return char == '"' || char == '\'' || char == '`'
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

func collapseWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
