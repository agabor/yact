package index

import "strings"

var kotlinSignatureRules = signatureRules{IsFunction: isKotlinFunctionLabel, IsPublic: isKotlinPublicLabel, SplitsAtLineBreak: true}

var kotlinModifiers = []string{"public", "private", "protected", "internal", "open", "final", "abstract", "override", "suspend", "inline", "infix", "operator", "tailrec", "external", "data", "sealed", "enum", "annotation", "companion", "lateinit", "const", "fun"}

var kotlinDeclarationKeywords = []string{"fun", "class", "object", "interface"}

var kotlinNonPublicModifiers = []string{"private", "protected", "internal"}

func isKotlinNameByte(char byte) bool {
	isLetter := (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
	isDigit := char >= '0' && char <= '9'
	return isLetter || isDigit || char == '_' || char == '.' || char == ':'
}

func skipToKotlinKeyword(source string, lineStart int) int {
	position := skipAtAttributes(source, skipIndentation(source, lineStart), isKotlinNameByte)
	for {
		if isKotlinDeclarationKeywordAt(source, position) {
			return position
		}
		modifier, matched := matchModifier(source, position, kotlinModifiers)
		if !matched {
			return position
		}
		position = skipAtAttributes(source, skipIndentation(source, position+len(modifier)), isKotlinNameByte)
	}
}

func isKotlinDeclarationKeywordAt(source string, position int) bool {
	return startsWithKeyword(source, position, kotlinDeclarationKeywords, " \t\r\n<(")
}

func isKotlinFunctionLabel(label string) bool {
	position := skipToKotlinKeyword(label, 0)
	if !startsWithKeyword(label, position, []string{"fun"}, " \t<(") {
		return false
	}
	return !strings.HasPrefix(label[position:], "fun interface ")
}

func isKotlinPublicLabel(label string) bool {
	position := skipAtAttributes(label, skipIndentation(label, 0), isKotlinNameByte)
	for {
		modifier, matched := matchModifier(label, position, kotlinModifiers)
		if !matched {
			return true
		}
		if containsString(kotlinNonPublicModifiers, modifier) {
			return false
		}
		position = skipAtAttributes(label, skipIndentation(label, position+len(modifier)), isKotlinNameByte)
	}
}
