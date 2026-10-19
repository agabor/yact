package commands

import "strings"

var phpModifiers = []string{"public", "protected", "private", "static", "abstract", "final", "readonly"}

var phpDeclarationKeywords = []string{"function", "class", "interface", "trait"}

var phpSignatureRules = signatureRules{
	IsDeclarationStart: isPHPDeclarationStart,
	IsMemberStart:      isPHPDeclarationStart,
	IsContainerStart:   isPHPContainerStart,
}

func isPHPDeclarationStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, phpModifiers)
	return startsWithKeyword(source, position, phpDeclarationKeywords, " &")
}

func isPHPContainerStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, phpModifiers)
	return startsWithKeyword(source, position, []string{"class", "interface", "trait"}, " \t\r\n")
}

func skipIndentation(source string, lineStart int) int {
	position := lineStart
	for position < len(source) && (source[position] == ' ' || source[position] == '\t') {
		position++
	}
	return position
}

func skipModifiers(source string, position int, modifiers []string) int {
	for {
		next, matched := skipOneModifier(source, position, modifiers)
		if !matched {
			return position
		}
		position = next
	}
}

func skipOneModifier(source string, position int, modifiers []string) (int, bool) {
	for _, modifier := range modifiers {
		if strings.HasPrefix(source[position:], modifier+" ") {
			return skipIndentation(source, position+len(modifier)), true
		}
	}
	return position, false
}

func startsWithKeyword(source string, position int, keywords []string, followers string) bool {
	for _, keyword := range keywords {
		if !strings.HasPrefix(source[position:], keyword) {
			continue
		}
		followerIndex := position + len(keyword)
		if followerIndex < len(source) && strings.IndexByte(followers, source[followerIndex]) >= 0 {
			return true
		}
	}
	return false
}