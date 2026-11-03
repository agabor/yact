package logic

import "strings"

// Separator is empty exactly when the text has no footer section. Footer may be empty while a separator is present.
type PromptSections struct {
	Body      string
	Separator string
	Footer    string
}

func SplitPromptFooter(promptText string) PromptSections {
	lines := strings.SplitAfter(promptText, "\n")
	offset := 0
	for _, line := range lines {
		if isPromptSeparatorLine(line) {
			return PromptSections{
				Body:      promptText[:offset],
				Separator: strings.TrimRight(line, "\r\n"),
				Footer:    promptText[offset+len(line):],
			}
		}
		offset += len(line)
	}
	return PromptSections{Body: promptText}
}

func JoinPromptSections(sections PromptSections) string {
	if sections.Separator == "" {
		return sections.Body + sections.Footer
	}
	return ensureTrailingNewline(sections.Body) + sections.Separator + "\n" + sections.Footer
}

func isPromptSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed != "" && strings.Trim(trimmed, "=") == ""
}

func ensureTrailingNewline(text string) string {
	if text == "" || strings.HasSuffix(text, "\n") {
		return text
	}
	return text + "\n"
}