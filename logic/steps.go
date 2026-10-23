package logic

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type PlanStep struct {
	Number      int
	Title       string
	Checked     bool
	HeadingLine int
	Text        string
}

type ParsedPlan struct {
	Preamble string
	Steps    []PlanStep
}

type StepRange struct {
	Start int
	End   int
}

type StepSelector struct {
	Next   bool
	Ranges []StepRange
}

type StepSelection struct {
	Steps    []PlanStep
	Warnings []string
	AllDone  bool
}

var checkedStepHeadingPattern = regexp.MustCompile(`^###\s+\[( |x|X)\]\s+(\d+)\.\s*(.*)$`)

var plainStepHeadingPattern = regexp.MustCompile(`^###\s+(\d+)\.\s*(.*)$`)

var stepCheckboxPattern = regexp.MustCompile(`^(###\s+\[)[ xX](\]\s+\d+\.)`)

var stepHeadingPrefixPattern = regexp.MustCompile(`^(###\s+)`)

func trimLineEnding(line string) string {
	return strings.TrimRight(line, "\r\n")
}

func parseStepHeading(line string) (int, string, bool, bool) {
	trimmedLine := trimLineEnding(line)

	if match := checkedStepHeadingPattern.FindStringSubmatch(trimmedLine); match != nil {
		number, err := strconv.Atoi(match[2])
		if err != nil {
			return 0, "", false, false
		}
		return number, match[3], match[1] != " ", true
	}

	if match := plainStepHeadingPattern.FindStringSubmatch(trimmedLine); match != nil {
		number, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, "", false, false
		}
		return number, match[2], false, true
	}

	return 0, "", false, false
}

func isStepBoundary(line string) bool {
	if _, _, _, isHeading := parseStepHeading(line); isHeading {
		return true
	}
	return strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ")
}

func findStepEnd(lines []string, headingLine int) int {
	end := headingLine + 1
	for end < len(lines) && !isStepBoundary(lines[end]) {
		end++
	}
	return end
}

func ParsePlan(plan string) ParsedPlan {
	lines := strings.SplitAfter(plan, "\n")
	parsedPlan := ParsedPlan{Preamble: plan}
	firstHeadingFound := false

	for lineIndex, line := range lines {
		number, title, checked, isHeading := parseStepHeading(line)
		if !isHeading {
			continue
		}

		if !firstHeadingFound {
			firstHeadingFound = true
			parsedPlan.Preamble = strings.Join(lines[:lineIndex], "")
		}

		end := findStepEnd(lines, lineIndex)
		parsedPlan.Steps = append(parsedPlan.Steps, PlanStep{
			Number:      number,
			Title:       title,
			Checked:     checked,
			HeadingLine: lineIndex,
			Text:        strings.Join(lines[lineIndex:end], ""),
		})
	}

	return parsedPlan
}

func (p ParsedPlan) Validate() error {
	seenNumbers := make(map[int]bool)
	for _, step := range p.Steps {
		if seenNumbers[step.Number] {
			return fmt.Errorf("the plan contains duplicate step number %d", step.Number)
		}
		seenNumbers[step.Number] = true
	}
	return nil
}

func (p ParsedPlan) AllChecked() bool {
	if len(p.Steps) == 0 {
		return false
	}
	for _, step := range p.Steps {
		if !step.Checked {
			return false
		}
	}
	return true
}

func (p ParsedPlan) UncheckedSteps() []PlanStep {
	uncheckedSteps := []PlanStep{}
	for _, step := range p.Steps {
		if !step.Checked {
			uncheckedSteps = append(uncheckedSteps, step)
		}
	}
	return uncheckedSteps
}

func parseStepNumber(text string) (int, error) {
	trimmedText := strings.TrimSpace(text)
	number, err := strconv.Atoi(trimmedText)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, fmt.Errorf("step number '%s' is too large", trimmedText)
		}
		return 0, fmt.Errorf("invalid step selector item '%s'", trimmedText)
	}
	if number < 1 {
		return 0, fmt.Errorf("step number must be at least 1, got %d", number)
	}
	return number, nil
}

func parseStepRange(item string) (StepRange, error) {
	parts := strings.Split(item, "-")
	if len(parts) > 2 {
		return StepRange{}, fmt.Errorf("invalid step selector item '%s'", item)
	}

	start, err := parseStepNumber(parts[0])
	if err != nil {
		return StepRange{}, err
	}

	if len(parts) == 1 {
		return StepRange{Start: start, End: start}, nil
	}

	end, err := parseStepNumber(parts[1])
	if err != nil {
		return StepRange{}, err
	}

	if start > end {
		return StepRange{}, fmt.Errorf("invalid step range '%s': start is greater than end", item)
	}

	return StepRange{Start: start, End: end}, nil
}

func isNextKeyword(item string) bool {
	return strings.EqualFold(strings.TrimSpace(item), "next")
}

func ParseStepSelector(selector string) (StepSelector, error) {
	if strings.TrimSpace(selector) == "" {
		return StepSelector{}, fmt.Errorf("steps selector is empty")
	}

	items := strings.Split(selector, ",")

	if len(items) == 1 && isNextKeyword(items[0]) {
		return StepSelector{Next: true}, nil
	}

	stepSelector := StepSelector{}
	for _, item := range items {
		trimmedItem := strings.TrimSpace(item)
		if trimmedItem == "" {
			return StepSelector{}, fmt.Errorf("steps selector contains an empty item")
		}
		if isNextKeyword(trimmedItem) {
			return StepSelector{}, fmt.Errorf("'next' cannot be combined with other step selectors")
		}

		stepRange, err := parseStepRange(trimmedItem)
		if err != nil {
			return StepSelector{}, err
		}
		stepSelector.Ranges = append(stepSelector.Ranges, stepRange)
	}

	return stepSelector, nil
}

func formatStepRange(stepRange StepRange) string {
	if stepRange.Start == stepRange.End {
		return strconv.Itoa(stepRange.Start)
	}
	return fmt.Sprintf("%d-%d", stepRange.Start, stepRange.End)
}

func FormatNumbers(numbers []int) string {
	texts := make([]string, 0, len(numbers))
	for _, number := range numbers {
		texts = append(texts, strconv.Itoa(number))
	}
	return strings.Join(texts, ", ")
}

func countStepsInRange(steps []PlanStep, stepRange StepRange) int {
	count := 0
	for _, step := range steps {
		if step.Number >= stepRange.Start && step.Number <= stepRange.End {
			count++
		}
	}
	return count
}

func validateRanges(plan ParsedPlan, ranges []StepRange) error {
	for _, stepRange := range ranges {
		expectedCount := stepRange.End - stepRange.Start + 1
		if countStepsInRange(plan.Steps, stepRange) != expectedCount {
			return fmt.Errorf("step range %s does not match the plan, existing steps: %s", formatStepRange(stepRange), FormatNumbers(StepNumbers(plan.Steps)))
		}
	}
	return nil
}

func isInAnyRange(number int, ranges []StepRange) bool {
	for _, stepRange := range ranges {
		if number >= stepRange.Start && number <= stepRange.End {
			return true
		}
	}
	return false
}

func selectNextStep(plan ParsedPlan) StepSelection {
	for _, step := range plan.Steps {
		if !step.Checked {
			return StepSelection{Steps: []PlanStep{step}}
		}
	}
	return StepSelection{AllDone: true}
}

func selectStepsInRanges(plan ParsedPlan, ranges []StepRange) StepSelection {
	selection := StepSelection{}
	for _, step := range plan.Steps {
		if !isInAnyRange(step.Number, ranges) {
			continue
		}
		selection.Steps = append(selection.Steps, step)
		if step.Checked {
			selection.Warnings = append(selection.Warnings, fmt.Sprintf("step %d is already done", step.Number))
		}
	}
	return selection
}

func SelectSteps(plan ParsedPlan, selector StepSelector) (StepSelection, error) {
	if err := plan.Validate(); err != nil {
		return StepSelection{}, err
	}

	if len(plan.Steps) == 0 {
		return StepSelection{}, fmt.Errorf("the plan contains no steps")
	}

	if selector.Next {
		return selectNextStep(plan), nil
	}

	if err := validateRanges(plan, selector.Ranges); err != nil {
		return StepSelection{}, err
	}

	return selectStepsInRanges(plan, selector.Ranges), nil
}

func BuildStepsPrompt(plan ParsedPlan, steps []PlanStep) string {
	var builder strings.Builder
	builder.WriteString(plan.Preamble)
	for _, step := range steps {
		builder.WriteString(step.Text)
	}
	return builder.String()
}

func markLineDone(line string) string {
	if stepCheckboxPattern.MatchString(line) {
		return stepCheckboxPattern.ReplaceAllString(line, "${1}x${2}")
	}
	return stepHeadingPrefixPattern.ReplaceAllString(line, "${1}[x] ")
}

func MarkStepsDone(plan string, steps []PlanStep) string {
	lines := strings.SplitAfter(plan, "\n")
	for _, step := range steps {
		if step.HeadingLine >= 0 && step.HeadingLine < len(lines) {
			lines[step.HeadingLine] = markLineDone(lines[step.HeadingLine])
		}
	}
	return strings.Join(lines, "")
}

func StepNumbers(steps []PlanStep) []int {
	numbers := make([]int, 0, len(steps))
	for _, step := range steps {
		numbers = append(numbers, step.Number)
	}
	return numbers
}