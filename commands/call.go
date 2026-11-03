package commands

import (
	"fmt"
	"os"

	"yact/config"
	"yact/logic"
)

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, prompt string, skills []logic.Skill) error {
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

	planPrompt := ""
	combinedPrompt := ""
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

	// Writing the plan is how a blocked prompt gets unblocked, so it is never gated.
	if !flags.WritePlan {
		gateText := prompt
		if planUsed {
			gateText = planPrompt
		}
		if err := logic.PromptGateError(logic.DecidePromptGate(gateText)); err != nil {
			return err
		}
	}

	requestContentForLLM, err := logic.LoadRequestContent()
	if err != nil {
		return err
	}
	var sentSteps []logic.PlanStep

	if planUsed {
		selection, err := selectPlanSteps(planPrompt, flags.StepsSelector())

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

		combinedPrompt = combinePromptText(promptFileText, prompt)
		requestContentForLLM.Prompt = combinedPrompt
	}
	requestContentForLLM, err = applyPromptLinks(requestContentForLLM)
	if err != nil {
		return err
	}
	requestContentForLLM.Skills = skills

	fileCounts, err := countFileLines(requestContentForLLM)
	if err != nil {
		return err
	}

	printWarnings(longFileWarnings(fileCounts))

	if flags.Verbose {
		printSentFiles(requestContentForLLM)
		printSentPrompt(requestContentForLLM.Prompt)
	}

	if flags.Dry {
		fmt.Println("Dry run: LLM call skipped.")
		return nil
	}

	// Later calls with a prompt read this file back, so it must hold the prompt that was sent.
	if !planUsed {
		if err := os.WriteFile(config.GetProjectPromptPath(), []byte(combinedPrompt), 0644); err != nil {
			return err
		}
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
		return err
	}

	// A step whose output was only printed or written as a plan stays unchecked.
	if planUsed && (mode == config.OutputApply || mode == config.OutputValidateAndApply) {
		return tickPlanSteps(planPrompt, sentSteps)
	}

	return nil
}
