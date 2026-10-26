package index

import "strings"

var jsModifiers = []string{"export", "default", "async"}

var jsMethodModifiers = []string{"static", "async"}

var jsControlKeywords = []string{"if", "for", "while", "switch", "catch", "with", "else", "do", "try", "finally", "return"}

var jsSignatureRules = signatureRules{IsFunction: isJSFunctionLabel, IsPublic: isJSPublicLabel, SplitsAtLineBreak: true}

func isJSFunctionLabel(label string) bool {
	position := skipModifiers(label, 0, jsModifiers)
	if startsWithKeyword(label, position, []string{"function"}, " *(") {
		return true
	}
	return isJSMethodLabel(label) || isJSArrowBinding(label)
}

func isJSArrowBinding(label string) bool {
	position := skipModifiers(label, 0, jsModifiers)
	return startsWithKeyword(label, position, []string{"const", "let", "var"}, " ") &&
		strings.Contains(label, "=") && strings.HasSuffix(label, "=>")
}

func isJSMethodLabel(label string) bool {
	namePosition := jsMethodNamePosition(label)
	nameEnd := skipJSIdentifier(label, namePosition)
	if nameEnd == namePosition || containsString(jsControlKeywords, label[namePosition:nameEnd]) {
		return false
	}
	position := skipIndentation(label, nameEnd)
	return position < len(label) && label[position] == '('
}

func jsMethodNamePosition(label string) int {
	position := skipModifiers(label, 0, jsMethodModifiers)
	if position < len(label) && label[position] == '*' {
		position++
	}
	return position
}

func isJSPublicLabel(label string) bool {
	if isJSMethodLabel(label) {
		return !strings.HasPrefix(label[jsMethodNamePosition(label):], "#")
	}
	return startsWithKeyword(label, 0, []string{"export"}, " ")
}

func skipJSIdentifier(source string, position int) int {
	for position < len(source) && isJSIdentifierChar(source[position]) {
		position++
	}
	return position
}

func isJSIdentifierChar(char byte) bool {
	return char == '_' || char == '$' || char == '#' ||
		(char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}
