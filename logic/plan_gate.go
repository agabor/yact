package logic

import (
	"errors"
	"strings"
)

type PromptGate int

const (
	PromptGateOpen PromptGate = iota
	PromptGateMissingContext
	PromptGateNoStepsRemain
	PromptGateUnansweredAssumptions
	PromptGateUnappliedAnswers
)

const missingContextHeading = "## Missing context"

func HasMissingContextSection(promptText string) bool {
	for _, line := range strings.Split(promptText, "\n") {
		if trimLineEnding(line) == missingContextHeading {
			return true
		}
	}
	return false
}

func hasUnansweredAssumption(assumptions []Assumption) bool {
	for _, assumption := range assumptions {
		if assumption.IsUnanswered() {
			return true
		}
	}
	return false
}

func hasUnappliedAnswer(assumptions []Assumption) bool {
	for _, assumption := range assumptions {
		if assumption.HasUnappliedAnswer() {
			return true
		}
	}
	return false
}

// A missing-context section takes precedence because the plan cannot be run
// until the files it asks for are supplied, even if all its steps are ticked.
func DecidePromptGate(promptText string) PromptGate {
	if HasMissingContextSection(promptText) {
		return PromptGateMissingContext
	}
	plan := ParsePlan(promptText)
	if len(plan.Steps) > 0 && plan.AllChecked() {
		return PromptGateNoStepsRemain
	}
	assumptions := ParseAssumptions(promptText)
	if hasUnansweredAssumption(assumptions) {
		return PromptGateUnansweredAssumptions
	}
	if hasUnappliedAnswer(assumptions) {
		return PromptGateUnappliedAnswers
	}
	return PromptGateOpen
}

func PromptGateError(gate PromptGate) error {
	switch gate {
	case PromptGateMissingContext:
		return errors.New("Files needed as context are missing; run with --plan to update the plan")
	case PromptGateNoStepsRemain:
		return errors.New("no steps remain; run with --plan to update the plan")
	case PromptGateUnansweredAssumptions:
		return errors.New("Assumptions are still open; tick the box of each one you accept, or write \"Answer: ...\" under it and run with --plan to update the plan")
	case PromptGateUnappliedAnswers:
		return errors.New("Answers under assumptions are not applied yet; run with --plan to update the plan")
	default:
		return nil
	}
}
