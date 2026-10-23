package commands

import "strings"

var phpModifiers = []string{"public", "protected", "private", "static", "abstract", "final", "readonly"}

var phpNonPublicModifiers = []string{"private", "protected"}

var phpDeclarationKeywords = []string{"function", "class", "interface", "trait"}

var phpSignatureRules = signatureRules{
	IsDeclarationStart:  isPHPDeclarationStart,
	IsMemberStart:       isPHPDeclarationStart,
	IsContainerStart:    isPHPContainerStart,
	IsPublicDeclaration: isPHPPublic,
	IsPublicMember:      isPHPPublic,
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

func isPHPPublic(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	for {
		modifier, matched := matchModifier(source, position, phpModifiers)
		if !matched {
			return true
		}
		if containsString(phpNonPublicModifiers, modifier) {
			return false
		}
		position = skipIndentation(source, position+len(modifier))
	}
}

func matchModifier(source string, position int, modifiers []string) (string, bool) {
	for _, modifier := range modifiers {
		if strings.HasPrefix(source[position:], modifier+" ") {
			return modifier, true
		}
	}
	return "", false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func skipIndentation(source string, lineStart int) int {
	position := lineStart
	for position < len(source) && (source[position] == ' ' || source[position] == '\t') {
		position++
	}
	return position
}

func skipAtAttributes(source string, position int, isNameByte func(char byte) bool) int {
	for position < len(source) && source[position] == '@' {
		position++
		for position < len(source) && isNameByte(source[position]) {
			position++
		}
		position = skipIndentation(source, position)
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
	modifier, matched := matchModifier(source, position, modifiers)
	if !matched {
		return position, false
	}
	return skipIndentation(source, position+len(modifier)), true
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