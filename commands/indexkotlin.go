package commands

var kotlinSignatureRules = signatureRules{
	IsDeclarationStart:  isKotlinDeclarationStart,
	IsMemberStart:       isKotlinDeclarationStart,
	IsContainerStart:    isKotlinContainerStart,
	IsPublicDeclaration: isKotlinPublic,
	IsPublicMember:      isKotlinPublic,
	EndsAtLineBreak:     true,
}

var kotlinModifiers = []string{"public", "private", "protected", "internal", "open", "final", "abstract", "override", "suspend", "inline", "infix", "operator", "tailrec", "external", "data", "sealed", "enum", "annotation", "companion", "lateinit", "const", "fun"}

var kotlinDeclarationKeywords = []string{"fun", "class", "object", "interface"}

var kotlinContainerKeywords = []string{"class", "object", "interface"}

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

func isKotlinDeclarationStart(source string, lineStart int) bool {
	return isKotlinDeclarationKeywordAt(source, skipToKotlinKeyword(source, lineStart))
}

func isKotlinContainerStart(source string, lineStart int) bool {
	position := skipToKotlinKeyword(source, lineStart)
	return startsWithKeyword(source, position, kotlinContainerKeywords, " \t\r\n<(")
}

func isKotlinPublic(source string, lineStart int) bool {
	position := skipAtAttributes(source, skipIndentation(source, lineStart), isKotlinNameByte)
	for {
		modifier, matched := matchModifier(source, position, kotlinModifiers)
		if !matched {
			return true
		}
		if containsString(kotlinNonPublicModifiers, modifier) {
			return false
		}
		position = skipAtAttributes(source, skipIndentation(source, position+len(modifier)), isKotlinNameByte)
	}
}