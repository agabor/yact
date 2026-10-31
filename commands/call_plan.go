package commands

import (
	"fmt"
	"strings"

	"yact/logic"
)

type PlanSelection struct {
	Prompt   string
	Steps    []logic.PlanStep
	AllDone  bool
	Warnings []string
}

func selectPlanSteps(planText string, selector string) (PlanSelection, error) {
	if selector == "" {
		parsedPlan := logic.ParsePlan(planText)
		remainingSteps := parsedPlan.UncheckedSteps()
		if len(parsedPlan.Steps) == len(remainingSteps) {
			return PlanSelection{Prompt: planText, Steps: parsedPlan.Steps}, nil
		}
		if parsedPlan.AllChecked() {
			return PlanSelection{AllDone: true}, nil
		}
		return PlanSelection{
			Prompt: logic.BuildStepsPrompt(parsedPlan, remainingSteps),
			Steps:  remainingSteps,
		}, nil
	}

	stepSelector, err := logic.ParseStepSelector(selector)
	if err != nil {
		return PlanSelection{}, err
	}
	if stepSelector.Kind == logic.StepSelectorWholePlan {
		return PlanSelection{Prompt: planText}, nil
	}

	parsedPlan := logic.ParsePlan(planText)
	selection, err := logic.SelectSteps(parsedPlan, stepSelector)
	if err != nil {
		return PlanSelection{}, err
	}

	if selection.AllDone {
		return PlanSelection{AllDone: true, Warnings: selection.Warnings}, nil
	}

	return PlanSelection{
		Prompt:   logic.BuildStepsPrompt(parsedPlan, selection.Steps),
		Steps:    selection.Steps,
		Warnings: selection.Warnings,
	}, nil
}

func tickPlanSteps(planText string, steps []logic.PlanStep) error {
	if len(steps) == 0 {
		return nil
	}

	if err := logic.WritePlan(logic.MarkStepsDone(planText, steps)); err != nil {
		return err
	}

	fmt.Printf("Marked done: %s\n", logic.FormatNumbers(logic.StepNumbers(steps)))
	return nil
}

func combinePromptText(promptFileText string, cliPrompt string) string {
	fileText := strings.TrimSpace(promptFileText)
	cliText := strings.TrimSpace(cliPrompt)
	if fileText == "" {
		return cliText
	}
	if cliText == "" {
		return fileText
	}
	return fileText + "\n\n" + cliText
}

func applyPromptLinks(requestContent logic.RequestContent) (logic.RequestContent, error) {
	if requestContent.Prompt == "" {
		return requestContent, nil
	}

	links, err := logic.ResolvePromptLinks(requestContent.Prompt)
	if err != nil {
		return requestContent, err
	}

	requestContent.Files = append([]string{}, links.Paths...)
	return requestContent, nil
}