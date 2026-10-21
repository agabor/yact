package commands

var swiftModifiers = []string{"public", "open", "internal", "fileprivate", "private", "static", "class", "final", "override", "mutating", "nonmutating", "required", "convenience", "dynamic", "lazy", "weak", "unowned"}

var swiftDeclarationKeywords = []string{"func", "init", "class", "struct", "enum", "protocol", "extension"}

var swiftContainerKeywords = []string{"class", "struct", "enum", "protocol", "extension"}

var swiftFunctionKeywords = []string{"func", "init"}

var swiftPublicModifiers = []string{"public", "open"}

const swiftKeywordFollowers = " \t\r\n(<?!"

var swiftSignatureRules = signatureRules{
	IsDeclarationStart:  isSwiftDeclarationStart,
	IsMemberStart:       isSwiftDeclarationStart,
	IsContainerStart:    isSwiftContainerStart,
	IsPublicDeclaration: isSwiftPublic,
	IsPublicMember:      isSwiftPublic,
	EndsAtLineBreak:     true,
}

func isSwiftIdentifierByte(char byte) bool {
	return char == '_' ||
		(char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

func skipSwiftAttributes(source string, position int) int {
	for position < len(source) && source[position] == '@' {
		position++
		for position < len(source) && isSwiftIdentifierByte(source[position]) {
			position++
		}
		position = skipIndentation(source, position)
	}
	return position
}

func isSwiftDeclarationStart(source string, lineStart int) bool {
	position := skipSwiftAttributes(source, skipIndentation(source, lineStart))
	return matchesSwiftKeywordAfterModifiers(source, position, swiftDeclarationKeywords, nil)
}

func isSwiftContainerStart(source string, lineStart int) bool {
	position := skipSwiftAttributes(source, skipIndentation(source, lineStart))
	return matchesSwiftKeywordAfterModifiers(source, position, swiftContainerKeywords, swiftFunctionKeywords)
}

func matchesSwiftKeywordAfterModifiers(source string, position int, keywords []string, blockingKeywords []string) bool {
	for {
		if startsWithKeyword(source, position, blockingKeywords, swiftKeywordFollowers) {
			return false
		}
		if startsWithKeyword(source, position, keywords, swiftKeywordFollowers) {
			return true
		}
		next, matched := skipOneModifier(source, position, swiftModifiers)
		if !matched {
			return false
		}
		position = skipSwiftAttributes(source, next)
	}
}

func isSwiftPublic(source string, lineStart int) bool {
	position := skipSwiftAttributes(source, skipIndentation(source, lineStart))
	for {
		modifier, matched := matchModifier(source, position, swiftModifiers)
		if !matched {
			return false
		}
		if containsString(swiftPublicModifiers, modifier) {
			return true
		}
		position = skipSwiftAttributes(source, skipIndentation(source, position+len(modifier)))
	}
}