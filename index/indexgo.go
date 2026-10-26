package index

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var goSignatureRules = signatureRules{IsFunction: isGoFunctionLabel, IsPublic: isGoPublicLabel, SplitsAtLineBreak: true}

func isGoFunctionLabel(label string) bool {
	return strings.HasPrefix(label, "func ") || strings.HasPrefix(label, "func(")
}

func isGoPublicLabel(label string) bool {
	position := skipIndentation(label, len("func"))
	if position < len(label) && label[position] == '(' {
		position = skipIndentation(label, skipGoReceiver(label, position))
	}
	if position >= len(label) {
		return false
	}
	first, _ := utf8.DecodeRuneInString(label[position:])
	return unicode.IsUpper(first)
}

func skipGoReceiver(label string, position int) int {
	closing := strings.IndexByte(label[position:], ')')
	if closing < 0 {
		return len(label)
	}
	return position + closing + 1
}
