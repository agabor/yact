package logic

import "strings"

// ReadLinks is the read link block above the separator and Body the text below it. The separator
// is not stored: JoinPromptSections writes it whenever ReadLinks is set.
type PromptSections struct {
	ReadLinks string
	Body      string
}

const promptSeparator = "-----"

func SplitPromptFooter(promptText string) PromptSections {
	lines := strings.SplitAfter(promptText, "\n")
	offset := 0
	for _, line := range lines {
		if isPromptSeparatorLine(line) {
			// Blank lines around the separator are layout, not content, so they are dropped here and restored by JoinPromptSections.
			return PromptSections{
				ReadLinks: trimBlankLinesAtEnd(promptText[:offset]),
				Body:      trimBlankLinesAtStart(promptText[offset+len(line):]),
			}
		}
		offset += len(line)
	}
	return PromptSections{Body: promptText}
}

// Blank lines are written around the separator for readability; SplitPromptFooter accepts text with or without them.
func JoinPromptSections(sections PromptSections) string {
	if sections.ReadLinks == "" {
		return sections.Body
	}
	return ensureTrailingNewline(sections.ReadLinks) + "\n" + promptSeparator + "\n\n" + sections.Body
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
