package commands

import (
	"fmt"
	"strings"

	"yact/config"
	"yact/logic"
	"yact/projectpath"
)

func storeBuffer(response string) error {
	if err := logic.PutBuffer(response); err != nil {
		return fmt.Errorf("error writing buffer: %w", err)
	}
	return nil
}

func formatPlanSummary(plan string) string {
	assumptions := len(logic.ParseAssumptions(plan))
	missingContext := logic.CountMissingContextEntries(plan)
	return fmt.Sprintf("Assumptions: %d\nMissing context: %d", assumptions, missingContext)
}

func deliverResponse(mode config.OutputMode, response string) error {
	switch mode {
	case config.OutputWritePlan:
		previousPlan, _, err := logic.ReadPlan()
		if err != nil {
			return err
		}
		carriedPlan := logic.CarryAssumptionTicks(previousPlan, response)
		if err := logic.WritePlan(carriedPlan); err != nil {
			return err
		}
		fmt.Printf("Plan written to %s\n", config.GetProjectPromptPath())
		fmt.Println(formatPlanSummary(carriedPlan))
		return nil
	case config.OutputPrintOnly:
		fmt.Println("\n" + response)
		return nil
	case config.OutputApply:
		elements, err := logic.ParseCodeFilesDetailed(response)
		if err != nil {
			return err
		}
		return processResponse(elements)
	}
	return fmt.Errorf("unknown output mode %d", mode)
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
