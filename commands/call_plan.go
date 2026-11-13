package commands

import (
	"fmt"
	"strings"

	"yact/logic"
	"yact/projectpath"
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
	sections := logic.SplitPromptFooter(promptFileText)
	sections.Body = joinPromptBody(sections.Body, cliPrompt)
	return logic.JoinPromptSections(sections)
}

func joinPromptBody(fileBody string, cliPrompt string) string {
	fileText := strings.TrimSpace(fileBody)
	cliText := strings.TrimSpace(cliPrompt)
	if fileText == "" {
		return cliText
	}
	if cliText == "" {
		return fileText
	}
	return fileText + "\n\n" + cliText
}

// The read link block above the separator is never sent as text, so only the body is sent, with keyword markers removed.
func sentPromptText(promptText string) string {
	return logic.StripKeywordMarkers(logic.SplitPromptFooter(promptText).Body)
}

func parseProjectPaths(raws []string) ([]projectpath.Path, error) {
	var paths []projectpath.Path
	for _, raw := range raws {
		projectPath, err := projectpath.Parse(raw)
		if err != nil {
			return nil, err
		}
		paths = append(paths, projectPath)
	}
	return paths, nil
}

func applyPromptLinks(requestContent logic.RequestContent) (logic.RequestContent, error) {
	if requestContent.Prompt == "" {
		return requestContent, nil
	}

	links, err := logic.ResolvePromptLinks(requestContent.Prompt)
	if err != nil {
		return requestContent, err
	}

	keywordPaths, err := resolvePromptKeywordFiles(requestContent.Prompt)
	if err != nil {
		return requestContent, err
	}

	paths := append([]string{}, links.Paths...)
	paths = append(paths, keywordPaths...)
	projectPaths, err := parseProjectPaths(logic.UniquePaths(paths))
	if err != nil {
		return requestContent, err
	}
	requestContent.Files = projectPaths
	requestContent.Prompt = sentPromptText(requestContent.Prompt)
	return requestContent, nil
}
