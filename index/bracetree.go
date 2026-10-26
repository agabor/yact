package index

import "strings"

type braceNode struct {
	Label    string
	Children []*braceNode
}

func parseBraceTree(source string, splitsAtLineBreak bool) *braceNode {
	root := &braceNode{}
	stack := []*braceNode{root}
	labelStart := 0
	parenDepth := 0
	index := 0
	for index < len(source) {
		switch {
		case strings.HasPrefix(source[index:], "//"):
			index = skipLineComment(source, index)
		case strings.HasPrefix(source[index:], "/*"):
			index = skipBlockComment(source, index)
		case isLiteralQuote(source[index]):
			index = skipLiteral(source, index)
		case source[index] == '\n':
			if splitsAtLineBreak && labelResetsAtNewline(source, index, parenDepth) {
				labelStart = index + 1
			}
			index++
		case source[index] == '(':
			parenDepth++
			index++
		case source[index] == ')':
			if parenDepth > 0 {
				parenDepth--
			}
			index++
		case source[index] == '{':
			node := &braceNode{Label: extractLabel(source, labelStart, index)}
			top := stack[len(stack)-1]
			top.Children = append(top.Children, node)
			stack = append(stack, node)
			labelStart = index + 1
			parenDepth = 0
			index++
		case source[index] == '}':
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
			labelStart = index + 1
			parenDepth = 0
			index++
		case source[index] == ';':
			labelStart = index + 1
			index++
		default:
			index++
		}
	}
	return root
}

func labelResetsAtNewline(source string, newline int, parenDepth int) bool {
	if parenDepth != 0 {
		return false
	}
	previous := previousNonWhitespaceIndex(source, newline)
	if previous >= 0 && strings.IndexByte(",(:=", source[previous]) >= 0 {
		return false
	}
	next := nextNonWhitespaceIndex(source, newline+1)
	return next >= len(source) || strings.IndexByte("{.:", source[next]) < 0
}

func previousNonWhitespaceIndex(source string, position int) int {
	for index := position - 1; index >= 0; index-- {
		if !isWhitespaceByte(source[index]) {
			return index
		}
	}
	return -1
}

func nextNonWhitespaceIndex(source string, position int) int {
	index := position
	for index < len(source) && isWhitespaceByte(source[index]) {
		index++
	}
	return index
}

func isWhitespaceByte(char byte) bool {
	return char == ' ' || char == '\t' || char == '\r' || char == '\n'
}

func extractLabel(source string, start int, end int) string {
	var builder strings.Builder
	index := start
	for index < end {
		switch {
		case strings.HasPrefix(source[index:], "//"):
			index = skipLineComment(source, index)
		case strings.HasPrefix(source[index:], "/*"):
			index = skipBlockComment(source, index)
		case isLiteralQuote(source[index]):
			literalEnd := skipLiteral(source, index)
			builder.WriteString(source[index:literalEnd])
			index = literalEnd
		default:
			builder.WriteByte(source[index])
			index++
		}
	}
	return collapseWhitespace(builder.String())
}
