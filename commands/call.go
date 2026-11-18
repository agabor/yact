package commands

import (
	"fmt"

	"yact/config"
	"yact/logic"
)

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string) error {
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

	planPrompt, planFound, err := logic.ReadPlan()
	if err != nil {
		return err
	}
	if !planFound {
		return fmt.Errorf("no plan found in %s, create a plan with --plan", config.GetProjectPromptPath())
	}

	// Writing the plan is how a blocked prompt gets unblocked, so it is never gated.
	if mode != config.OutputWritePlan {
		if err := logic.PromptGateError(logic.DecidePromptGate(planPrompt)); err != nil {
			return err
		}
	}

	requestContentForLLM, err := logic.LoadRequestContent()
	if err != nil {
		return err
	}

	requestContentForLLM.Prompt = planPrompt
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
