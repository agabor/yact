package logic

import "strings"

type GoalState int

const (
	GoalUnchanged GoalState = iota
	GoalChanged
	GoalNew
)

type PlanChanges struct {
	Goal         GoalState
	NewSteps     int
	ChangedSteps int
	RemovedSteps int
}

const (
	goalHeading  = "# Goal"
	stepsHeading = "## Steps"
)

// Steps are matched by title, not by number, so renumbering a plan is not reported as a change.
func ComparePlans(previousPlan string, newPlan string) PlanChanges {
	previousKeys := stepKeysByTitle(splitStepBlocks(sectionLines(previousPlan, stepsHeading)))
	newBlocks := splitStepBlocks(sectionLines(newPlan, stepsHeading))

	changes := PlanChanges{Goal: compareGoal(previousPlan, newPlan)}
	newTitles := make(map[string]bool, len(newBlocks))
	for _, block := range newBlocks {
		title := stepTitle(block)
		newTitles[title] = true
		previousKey, existed := previousKeys[title]
		switch {
		case !existed:
			changes.NewSteps++
		case previousKey != stepKey(block):
			changes.ChangedSteps++
		}
	}
	for title := range previousKeys {
		if !newTitles[title] {
			changes.RemovedSteps++
		}
	}
	return changes
}

// Surrounding whitespace is layout, so only the trimmed goal text decides whether the goal changed.
func compareGoal(previousPlan string, newPlan string) GoalState {
	previousGoal := goalText(previousPlan)
	if previousGoal == "" {
		return GoalNew
	}
	if previousGoal != goalText(newPlan) {
		return GoalChanged
	}
	return GoalUnchanged
}

func goalText(plan string) string {
	return strings.TrimSpace(strings.Join(sectionLines(plan, goalHeading), "\n"))
}

// Only top-level headings end a section, so the "###" step headings stay inside the steps section.
func sectionLines(plan string, heading string) []string {
	lines := strings.Split(plan, "\n")
	start := -1
	for i, line := range lines {
		if start < 0 {
			if trimLineEnding(line) == heading {
				start = i + 1
			}
			continue
		}
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			return lines[start:i]
		}
	}
	if start < 0 {
		return nil
	}
	return lines[start:]
}

func splitStepBlocks(stepLines []string) []string {
	var blocks []string
	var current []string
	for _, line := range stepLines {
		if strings.HasPrefix(line, "### ") {
			if current != nil {
				blocks = append(blocks, strings.Join(current, "\n"))
			}
			current = []string{line}
			continue
		}
		if current != nil {
			current = append(current, line)
		}
	}
	if current != nil {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func stepTitle(block string) string {
	firstLine := strings.SplitN(block, "\n", 2)[0]
	heading := strings.TrimSpace(strings.TrimPrefix(firstLine, "### "))
	withoutNumber := strings.TrimLeft(heading, "0123456789")
	return strings.TrimSpace(strings.TrimPrefix(withoutNumber, "."))
}

// The body is trimmed so that blank lines around a step do not count as an edit.
func stepKey(block string) string {
	body := ""
	if parts := strings.SplitN(block, "\n", 2); len(parts) == 2 {
		body = parts[1]
	}
	return stepTitle(block) + "\n" + strings.TrimSpace(body)
}

func stepKeysByTitle(blocks []string) map[string]string {
	keys := make(map[string]string, len(blocks))
	for _, block := range blocks {
		keys[stepTitle(block)] = stepKey(block)
	}
	return keys
}
