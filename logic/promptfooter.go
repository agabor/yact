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
			// Blank lines around the separator are layout, not content, so they are dropped here and restored by JoinPromptSections.
			return PromptSections{
				Body:      trimBlankLinesAtEnd(promptText[:offset]),
				Separator: strings.TrimRight(line, "\r\n"),
				Footer:    trimBlankLinesAtStart(promptText[offset+len(line):]),
			}
		}
		offset += len(line)
	}
	return PromptSections{Body: promptText}
}

// Blank lines are written around the separator for readability; SplitPromptFooter accepts text with or without them.
func JoinPromptSections(sections PromptSections) string {
	if sections.Separator == "" {
		return sections.Body + sections.Footer
	}
	text := ""
	if body := ensureTrailingNewline(sections.Body); body != "" {
		text = body + "\n"
	}
	text += sections.Separator + "\n"
	if sections.Footer != "" {
		text += "\n" + sections.Footer
	}
	return text
}

func isPromptSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed != "" && strings.Trim(trimmed, "-") == ""
}

func ensureTrailingNewline(text string) string {
	if text == "" || strings.HasSuffix(text, "\n") {
		return text
	}
	return text + "\n"
}

func trimBlankLinesAtEnd(text string) string {
	trimmed := strings.TrimRight(text, " \t\r\n")
	if trimmed == "" {
		return ""
	}
	return trimmed + "\n"
}

func trimBlankLinesAtStart(text string) string {
	for {
		lineEnd := strings.IndexByte(text, '\n')
		if lineEnd < 0 {
			if strings.TrimSpace(text) == "" {
				return ""
			}
			return text
		}
		if strings.TrimSpace(text[:lineEnd]) != "" {
			return text
		}
		text = text[lineEnd+1:]
	}
}
