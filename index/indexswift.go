package index

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

func isSwiftDeclarationStart(source string, lineStart int) bool {
	position := skipAtAttributes(source, skipIndentation(source, lineStart), isSwiftIdentifierByte)
	return matchesSwiftKeywordAfterModifiers(source, position, swiftDeclarationKeywords, nil)
}

func isSwiftContainerStart(source string, lineStart int) bool {
	position := skipAtAttributes(source, skipIndentation(source, lineStart), isSwiftIdentifierByte)
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
		position = skipAtAttributes(source, next, isSwiftIdentifierByte)
	}
}

func isSwiftPublic(source string, lineStart int) bool {
	position := skipAtAttributes(source, skipIndentation(source, lineStart), isSwiftIdentifierByte)
	for {
		modifier, matched := matchModifier(source, position, swiftModifiers)
		if !matched {
			return false
		}
		if containsString(swiftPublicModifiers, modifier) {
			return true
		}
		position = skipAtAttributes(source, skipIndentation(source, position+len(modifier)), isSwiftIdentifierByte)
	}
}
