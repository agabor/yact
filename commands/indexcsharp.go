package commands

import (
	"strings"
	"unicode"
)

var csharpModifiers = []string{
	"public", "private", "protected", "internal", "static", "abstract", "sealed",
	"virtual", "override", "async", "readonly", "partial", "extern", "unsafe",
	"new", "const", "volatile", "required",
}

var csharpTypeKeywords = []string{"class", "struct", "interface", "record", "enum"}

var csharpContainerKeywords = []string{"class", "struct", "interface", "record"}

var csharpControlKeywords = []string{"if", "for", "foreach", "while", "switch", "using", "lock", "return", "catch", "else", "throw", "await", "yield", "new"}

var csharpSignatureRules = signatureRules{
	IsDeclarationStart:  isCSharpTypeStart,
	IsMemberStart:       isCSharpMemberStart,
	IsContainerStart:    isCSharpContainerStart,
	IsPublicDeclaration: isCSharpPublic,
	IsPublicMember:      isCSharpPublic,
}

func isCSharpTypeStart(source string, lineStart int) bool {
	position := skipModifiers(source, skipIndentation(source, lineStart), csharpModifiers)
	return startsWithKeyword(source, position, csharpTypeKeywords, " \t\r\n")
}

func isCSharpContainerStart(source string, lineStart int) bool {
	position := skipModifiers(source, skipIndentation(source, lineStart), csharpModifiers)
	return startsWithKeyword(source, position, csharpContainerKeywords, " \t\r\n")
}

func isCSharpMethodStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	if containsString(csharpControlKeywords, readIdentifier(source, position)) {
		return false
	}
	position = skipModifiers(source, position, csharpModifiers)
	return hasCSharpMethodShape(restOfLine(source, position))
}

func isCSharpMemberStart(source string, lineStart int) bool {
	return isCSharpTypeStart(source, lineStart) || isCSharpMethodStart(source, lineStart)
}

func isCSharpPublic(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	for {
		modifier, matched := matchModifier(source, position, csharpModifiers)
		if !matched {
			return false
		}
		if modifier == "public" {
			return true
		}
		position = skipIndentation(source, position+len(modifier))
	}
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

func restOfLine(source string, position int) string {
	newline := strings.IndexByte(source[position:], '\n')
	if newline < 0 {
		return source[position:]
	}
	return source[position : position+newline]
}