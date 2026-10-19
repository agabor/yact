package commands

var jsModifiers = []string{"export", "default", "async"}

var jsDeclarationKeywords = []string{"function", "class"}

func scanJSSignatures(source string) []string {
	return scanSignatures(source, isJSDeclarationStart)
}

func isJSDeclarationStart(source string, lineStart int) bool {
	position := skipIndentation(source, lineStart)
	position = skipModifiers(source, position, jsModifiers)
	return startsWithKeyword(source, position, jsDeclarationKeywords, " *")
}