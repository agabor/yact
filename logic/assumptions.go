package logic

import (
	"regexp"
	"strings"
)

// Assumption is an entry of the Assumptions section. The person accepts it by
// ticking its box, or answers it by writing "Answer:" under it.
type Assumption struct {
	ID      string
	Text    string
	Checked bool
	Answer  string
}

// ItemLine is the index of the entry's first line in the lines of the prompt,
// so that its box can be rewritten in place.
type assumptionEntry struct {
	Value    Assumption
	ItemLine int
}

const assumptionsHeading = "## Assumptions"

const answerPrefix = "Answer:"

var assumptionIDPattern = regexp.MustCompile(`^\*\*A(\d+)\.`)

func ParseAssumptions(promptText string) []Assumption {
	entries := scanAssumptions(promptText)
	if len(entries) == 0 {
		return nil
	}
	assumptions := make([]Assumption, 0, len(entries))
	for _, entry := range entries {
		assumptions = append(assumptions, entry.Value)
	}
	return assumptions
}

// Ticks are carried only for entries whose id and text are unchanged, so that a
// reworded assumption is asked again instead of keeping an acceptance it no longer has.
func CarryAssumptionTicks(previousText string, newText string) string {
	previous := make(map[string]Assumption)
	for _, entry := range scanAssumptions(previousText) {
		if entry.Value.ID == "" {
			continue
		}
		if _, seen := previous[entry.Value.ID]; !seen {
			previous[entry.Value.ID] = entry.Value
		}
	}

	lines := strings.SplitAfter(newText, "\n")
	for _, entry := range scanAssumptions(newText) {
		lines[entry.ItemLine] = rewriteAssumptionBox(lines[entry.ItemLine], isCarriedTick(previous, entry.Value))
	}
	return strings.Join(lines, "")
}

func (a Assumption) IsUnanswered() bool {
	return !a.Checked && a.Answer == ""
}

func (a Assumption) HasUnappliedAnswer() bool {
	return !a.Checked && a.Answer != ""
}

func scanAssumptions(promptText string) []assumptionEntry {
	lines := strings.SplitAfter(promptText, "\n")
	start, end, found := assumptionSectionBounds(lines)
	if !found {
		return nil
	}

	var entries []assumptionEntry
	for itemLine := start; itemLine < end; {
		if !isAssumptionItemLine(lines[itemLine]) {
			itemLine++
			continue
		}
		entryEnd := assumptionEntryEnd(lines, itemLine, end)
		entries = append(entries, parseAssumptionEntry(lines, itemLine, entryEnd))
		itemLine = entryEnd
	}
	return entries
}

func assumptionSectionBounds(lines []string) (start int, end int, found bool) {
	for index, line := range lines {
		if !found {
			if trimLineEnding(line) == assumptionsHeading {
				start = index + 1
				found = true
			}
			continue
		}
		if strings.HasPrefix(line, "#") || isPromptSeparatorLine(line) {
			return start, index, true
		}
	}
	if !found {
		return 0, 0, false
	}
	return start, len(lines), true
}

func assumptionEntryEnd(lines []string, itemLine int, sectionEnd int) int {
	entryEnd := itemLine + 1
	for entryEnd < sectionEnd && isAssumptionContinuation(lines[entryEnd]) {
		entryEnd++
	}
	return entryEnd
}

func isAssumptionContinuation(line string) bool {
	content := trimLineEnding(line)
	return strings.TrimSpace(content) != "" &&
		!isAssumptionItemLine(line) &&
		!strings.HasPrefix(content, "#") &&
		!isPromptSeparatorLine(line)
}

func isAssumptionItemLine(line string) bool {
	_, isItem := assumptionItemBody(trimLineEnding(line))
	return isItem
}

func assumptionItemBody(content string) (string, bool) {
	trimmed := strings.TrimLeft(content, " ")
	if !strings.HasPrefix(trimmed, "- ") {
		return "", false
	}
	return trimmed[len("- "):], true
}

func hasAssumptionBox(body string) bool {
	return strings.HasPrefix(body, "[ ] ") || strings.HasPrefix(body, "[x] ") || strings.HasPrefix(body, "[X] ")
}

func parseAssumptionEntry(lines []string, itemLine int, entryEnd int) assumptionEntry {
	body, _ := assumptionItemBody(trimLineEnding(lines[itemLine]))
	checked := false
	if hasAssumptionBox(body) {
		checked = body[1] == 'x' || body[1] == 'X'
		body = body[len("[ ] "):]
	}

	textLines := []string{body}
	answerLines := []string{}
	inAnswer := false
	for _, line := range lines[itemLine+1 : entryEnd] {
		trimmed := strings.TrimSpace(line)
		if !inAnswer && strings.HasPrefix(trimmed, answerPrefix) {
			inAnswer = true
			trimmed = strings.TrimPrefix(trimmed, answerPrefix)
		}
		if inAnswer {
			answerLines = append(answerLines, trimmed)
		} else {
			textLines = append(textLines, trimmed)
		}
	}

	return assumptionEntry{
		Value: Assumption{
			ID:      assumptionID(body),
			Text:    normalizeAssumptionText(strings.Join(textLines, "\n")),
			Checked: checked,
			Answer:  strings.TrimSpace(strings.Join(answerLines, "\n")),
		},
		ItemLine: itemLine,
	}
}

func assumptionID(text string) string {
	match := assumptionIDPattern.FindStringSubmatch(text)
	if match == nil {
		return ""
	}
	return "A" + match[1]
}

func normalizeAssumptionText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func isCarriedTick(previous map[string]Assumption, entry Assumption) bool {
	before, found := previous[entry.ID]
	return found && before.Checked && before.Text == entry.Text
}

func rewriteAssumptionBox(line string, checked bool) string {
	content := trimLineEnding(line)
	body, isItem := assumptionItemBody(content)
	if !isItem || !hasAssumptionBox(body) {
		return line
	}
	boxIndex := len(content) - len(body) + 1
	mark := " "
	if checked {
		mark = "x"
	}
	return content[:boxIndex] + mark + line[boxIndex+1:]
}
