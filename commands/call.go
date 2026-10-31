package commands

import (
	"fmt"

	"yact/config"
	"yact/logic"
)

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, prompt string) error {
	mode, err := config.ParseOutputMode(flags)
	if err != nil {
		return err
	}

	if flags.Batch {
		if err := checkNoPendingBatch(); err != nil {
			return err
		}
	}

	if err := validateStepsOptions(flags.Steps, flags.WritePlan, prompt); err != nil {
		return err
	}

	if err := applyIndexOverride(flags); err != nil {
		return err
	}

	applyModelOverride(cfg, flags.ModelOverride())
	applyMaxInputLinesOverride(cfg, flags.MaxInputLines)

	planPrompt := ""
	planUsed := false

	if prompt == "" {
		var planFound bool
		planPrompt, planFound, err = logic.ReadPlan()
		if err != nil {
			return err
		}
		if !planFound {
			return fmt.Errorf("no plan found in %s, provide a prompt or create a plan with --plan", config.GetProjectPromptPath())
		}
		planUsed = true
	}

	requestContentForLLM, err := logic.LoadRequestContent()
	if err != nil {
		return err
	}
	var sentSteps []logic.PlanStep

	if planUsed {
		selection, err := selectPlanSteps(planPrompt, flags.Steps)

		if err != nil {
			return err
		}
		printWarnings(selection.Warnings)
		if selection.AllDone {
			fmt.Println("All plan steps are done.")
			return nil
		}
		if len(selection.Steps) > 0 {
			fmt.Printf("Plan steps: %s\n", logic.FormatNumbers(logic.StepNumbers(selection.Steps)))
		}
		sentSteps = selection.Steps
		requestContentForLLM.Prompt = selection.Prompt
	} else {
		promptFileText, _, err := logic.ReadPlan()
		if err != nil {
			return err
		}

		requestContentForLLM.Prompt = combinePromptText(promptFileText, prompt)
	}
	requestContentForLLM, err = applyPromptLinks(requestContentForLLM)
	if err != nil {
		return err
	}

	totalLength, fileCounts, err := countInputLines(requestContentForLLM)
	if err != nil {
		return err
	}

	if totalLength > cfg.MaxInputLines {
		return fmt.Errorf("input exceeds maximum number allowed lines (%d), got %d. Increase max_input_lines to override.", cfg.MaxInputLines, totalLength)
	}

	printWarnings(longFileWarnings(fileCounts))

	fmt.Printf("Input lines: %d\n", totalLength)

	if flags.Verbose {
		printSentFiles(requestContentForLLM)
		printSentPrompt(requestContentForLLM.Prompt)
	}

	if flags.Dry {
		fmt.Println("Dry run: LLM call skipped.")
		return nil
	}

	if flags.Batch {
		return submitBatch(requestContentForLLM, cfg, systemPrompt, mode)
	}

	result, err := callLLM(requestContentForLLM, flags.NoProgress, cfg, systemPrompt)
	if err != nil {
		return err
	}

	printCallResult(result)

	if err := storeBuffer(result.Text); err != nil {
		return err
	}

	if err := deliverResponse(mode, result.Text); err != nil {
		fmt.Printf("Input lines: %d\n", totalLength)
		return err
	}

	// A step whose output was only printed or written as a plan stays unchecked.
	if planUsed && (mode == config.OutputApply || mode == config.OutputValidateAndApply) {
		return tickPlanSteps(planPrompt, sentSteps)
	}

	return nil
}