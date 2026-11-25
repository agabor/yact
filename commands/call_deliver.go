package commands

import (
	"fmt"
	"strings"

	"yact/config"
	"yact/logic"
	"yact/projectpath"
)

func storeBuffer(session logic.Session) error {
	if err := logic.PutBuffer(session); err != nil {
		return fmt.Errorf("error writing buffer: %w", err)
	}
	return nil
}

func formatPlanSummary(plan string) string {
	assumptions := len(logic.ParseAssumptions(plan))
	missingContext := logic.CountMissingContextEntries(plan)
	return fmt.Sprintf("Assumptions: %d\nMissing context: %d", assumptions, missingContext)
}

func formatPlanChanges(changes logic.PlanChanges) string {
	return fmt.Sprintf(
		"Goal: %s\nSteps new: %d\nSteps changed: %d\nSteps removed: %d",
		goalStateName(changes.Goal),
		changes.NewSteps,
		changes.ChangedSteps,
		changes.RemovedSteps,
	)
}

func goalStateName(state logic.GoalState) string {
	switch state {
	case logic.GoalUnchanged:
		return "unchanged"
	case logic.GoalChanged:
		return "changed"
	case logic.GoalNew:
		return "new"
	}
	return fmt.Sprintf("unknown goal state %d", state)
}

func deliverResponse(session logic.Session) error {
	switch session.Mode {
	case config.OutputWritePlan:
		previousPlan, _, err := logic.ReadPlan()
		if err != nil {
			return err
		}
		carriedPlan := logic.CarryAssumptionTicks(previousPlan, session.Response)
		if err := logic.WritePlan(carriedPlan); err != nil {
			return err
		}
		fmt.Printf("Plan written to %s\n", config.GetProjectPromptPath())
		fmt.Println(formatPlanSummary(carriedPlan))
		fmt.Println(formatPlanChanges(logic.ComparePlans(previousPlan, carriedPlan)))
		return nil
	case config.OutputPrintOnly:
		fmt.Println("\n" + session.Response)
		return nil
	case config.OutputApply:
		elements, err := logic.ParseCodeFilesDetailed(session.Response, session.Linked)
		if err != nil {
			return err
		}
		return processResponse(elements)
	}
	return fmt.Errorf("unknown output mode %d", session.Mode)
}

func processResponse(elements []logic.ParsedElement) error {
	var parseErrors []string

	llmFeedback := ""

	for _, element := range elements {
		switch element.Kind {
		case logic.ElementCodeFile:
			v := element.CodeFile
			outcome, err := v.Apply()
			if err != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
				continue
			}
			printApplyOutcome(v.Path, outcome)
			if projectpath.Join(config.GetProjectTempDir(), "feedback.md") == v.Path {
				llmFeedback = v.Content
			}
		case logic.ElementRejectedFile:
			parseErrors = append(parseErrors, element.Text)
		case logic.ElementText:
			if strings.TrimSpace(element.Text) != "" {
				fmt.Println(element.Text)
			}
		}
	}

	if len(llmFeedback) > 0 {
		fmt.Println(llmFeedback)
	}

	if len(parseErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	return nil
}
