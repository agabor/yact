package index

var swiftModifiers = []string{"public", "open", "internal", "fileprivate", "private", "static", "class", "final", "override", "mutating", "nonmutating", "required", "convenience", "dynamic", "lazy", "weak", "unowned"}

var swiftFunctionKeywords = []string{"func", "init"}

var swiftPublicModifiers = []string{"public", "open"}

const swiftKeywordFollowers = " \t\r\n(<?!"

var swiftSignatureRules = signatureRules{IsFunction: isSwiftFunctionLabel, IsPublic: isSwiftPublicLabel, SplitsAtLineBreak: true}

func isSwiftIdentifierByte(char byte) bool {
	return char == '_' ||
		(char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

func isSwiftFunctionLabel(label string) bool {
	position := skipAtAttributes(label, 0, isSwiftIdentifierByte)
	return matchesSwiftKeywordAfterModifiers(label, position, swiftFunctionKeywords, nil)
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

func isSwiftPublicLabel(label string) bool {
	position := skipAtAttributes(label, 0, isSwiftIdentifierByte)
	for {
		modifier, matched := matchModifier(label, position, swiftModifiers)
		if !matched {
			return false
		}
		if containsString(swiftPublicModifiers, modifier) {
			return true
		}
		position = skipAtAttributes(label, skipIndentation(label, position+len(modifier)), isSwiftIdentifierByte)
	}
}
