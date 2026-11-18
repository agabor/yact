package commands

import (
	"fmt"
	"os"

	"yact/config"
	"yact/logic"
)

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, prompt string) error {
	mode := config.ParseOutputMode(flags)

	if flags.Batch {
		if err := checkNoPendingBatch(); err != nil {
			return err
		}
	}

	if err := applyIndexOverride(flags.Index); err != nil {
		return err
	}

	applyModelOverride(cfg, flags.ModelOverride())

	var err error
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
	if mode != config.OutputWritePlan {
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

	if planUsed {
		requestContentForLLM.Prompt = planPrompt
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

	fileCounts, err := countFileLines(requestContentForLLM)
	if err != nil {
		return err
	}

	printWarnings(longFileWarnings(fileCounts))

	if flags.Verbose {
		printSentFiles(requestContentForLLM)
		printSentPrompt(requestContentForLLM.Prompt)
	}

	if flags.DryRun {
		fmt.Println("Dry run: LLM call skipped.")
		return nil
	}

	// Later calls with a prompt read this file back, so it must hold the prompt that was sent.
	if !planUsed {
		if err := os.WriteFile(string(config.GetProjectPromptPath()), []byte(combinedPrompt), 0644); err != nil {
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

	return deliverResponse(mode, result.Text)
}
