package index

import (
	"strings"
	"unicode"
)

var csharpModifiers = []string{
	"public", "private", "protected", "internal", "static", "abstract", "sealed",
	"virtual", "override", "async", "readonly", "partial", "extern", "unsafe",
	"new", "const", "volatile", "required",
}

var csharpControlKeywords = []string{"if", "for", "foreach", "while", "switch", "using", "lock", "return", "catch", "else", "throw", "await", "yield", "new"}

var csharpSignatureRules = signatureRules{IsFunction: isCSharpFunctionLabel, IsPublic: isCSharpPublicLabel, CleanLabel: cleanCSharpLabel}

func isCSharpFunctionLabel(label string) bool {
	position := skipIndentation(label, 0)
	if containsString(csharpControlKeywords, readIdentifier(label, position)) {
		return false
	}
	afterModifiers := skipModifiers(label, position, csharpModifiers)
	if hasCSharpMethodShape(label[afterModifiers:]) {
		return true
	}
	return afterModifiers > position && isCSharpConstructorShape(label[afterModifiers:])
}

func isCSharpConstructorShape(text string) bool {
	name := readIdentifier(text, 0)
	return name != "" && strings.HasPrefix(text[len(name):], "(")
}

func isCSharpPublicLabel(label string) bool {
	position := skipIndentation(label, 0)
	for {
		modifier, matched := matchModifier(label, position, csharpModifiers)
		if !matched {
			return false
		}
		if modifier == "public" {
			return true
		}
		position = skipIndentation(label, position+len(modifier))
	}
}

func cleanCSharpLabel(label string) string {
	trimmed := strings.TrimSpace(label)
	for strings.HasPrefix(trimmed, "[") {
		closing := strings.IndexByte(trimmed, ']')
		if closing < 0 {
			break
		}
		trimmed = strings.TrimSpace(trimmed[closing+1:])
	}
	return trimmed
}

func hasCSharpMethodShape(line string) bool {
	parenthesis := strings.IndexByte(line, '(')
	if parenthesis < 0 {
		return false
	}
	prefix := strings.TrimSpace(line[:parenthesis])
	if strings.ContainsAny(prefix, "=;{}") {
		return false
	}
	separator := strings.LastIndexAny(prefix, " \t")
	if separator < 0 {
		return false
	}
	returnType := strings.TrimSpace(prefix[:separator])
	return returnType != "" && isCSharpMethodName(prefix[separator+1:])
}

func isCSharpMethodName(name string) bool {
	genericStart := strings.IndexByte(name, '<')
	if genericStart < 0 {
		return isIdentifier(name)
	}
	return isIdentifier(name[:genericStart]) && strings.HasSuffix(name, ">")
}

func isIdentifier(text string) bool {
	if text == "" || unicode.IsDigit([]rune(text)[0]) {
		return false
	}
	return strings.IndexFunc(text, func(char rune) bool { return !isIdentifierRune(char) }) < 0
}

func isIdentifierRune(char rune) bool {
	return char == '_' || unicode.IsLetter(char) || unicode.IsDigit(char)
}

func readIdentifier(source string, position int) string {
	end := position
	for end < len(source) && isIdentifierRune(rune(source[end])) {
		end++
	}
	return source[position:end]
}
