package index

var jsModifiers = []string{"export", "default", "async"}

var jsDeclarationKeywords = []string{"function", "class"}

var jsMethodModifiers = []string{"static", "async"}

var jsSignatureRules = signatureRules{
	IsDeclarationStart:  isJSDeclarationStart,
	IsMemberStart:       isJSMemberStart,
	IsContainerStart:    isJSContainerStart,
	IsPublicDeclaration: isJSPublicDeclaration,
	IsPublicMember:      isJSPublicMember,
}

func isJSDeclarationStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, jsModifiers)
	return startsWithKeyword(source, position, jsDeclarationKeywords, " *")
}

func isJSContainerStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, jsModifiers)
	return startsWithKeyword(source, position, []string{"class"}, " *")
}

func isJSMemberStart(source string, lineStart int) bool {
	return isJSDeclarationStart(source, lineStart) || isJSMethodStart(source, lineStart)
}

func isJSPublicDeclaration(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	return startsWithKeyword(source, position, []string{"export"}, " ")
}

func isJSPublicMember(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, jsMethodModifiers)
	if position < len(source) && source[position] == '*' {
		position = skipIndentation(source, position+1)
	}
	return position >= len(source) || source[position] != '#'
}

func isJSMethodStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, jsMethodModifiers)
	if position < len(source) && source[position] == '*' {
		position++
	}
	nameEnd := skipJSIdentifier(source, position)
	if nameEnd == position {
		return false
	}
	position = skipIndentation(source, nameEnd)
	return position < len(source) && source[position] == '('
}

func skipJSIdentifier(source string, position int) int {
	for position < len(source) && isJSIdentifierChar(source[position]) {
		position++
	}
	return position
}

func isJSIdentifierChar(char byte) bool {
	return char == '_' || char == '$' || char == '#' ||
		(char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}
