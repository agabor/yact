package commands

import (
	"fmt"

	"yact/api"
	"yact/config"
	"yact/logic"
)

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, newClient api.ClientFactory) error {
	mode := config.ParseOutputMode(flags)

	if flags.Batch {
		if err := checkNoPendingBatch(); err != nil {
			return err
		}
	}

	applyModelOverride(cfg, flags.Model)

	planPrompt, planFound, err := logic.ReadPlan()
	if err != nil {
		return err
	}
	if !planFound {
		return fmt.Errorf("no plan found in %s", config.GetProjectPromptPath())
	}

	// Writing the plan is how a blocked prompt gets unblocked, so it is never gated.
	if mode != config.OutputWritePlan {
		if err := logic.PromptGateError(logic.DecidePromptGate(planPrompt)); err != nil {
			return err
		}
	}

	// The index is synced only after the gate passes, so a blocked call leaves it untouched.
	if !flags.DryRun {
		if err := syncIndexFile(indexSettingOrNone(flags.Index)); err != nil {
			return err
		}
	}

	requestContentForLLM, err := logic.LoadRequestContent()
	if err != nil {
		return err
	}

	requestContentForLLM.Prompt = planPrompt
	err = applyPromptLinks(requestContentForLLM)
	if err != nil {
		return err
	}

	if flags.DryRun {
		fmt.Println("Dry run: LLM call skipped.")
		return nil
	}

	if flags.Batch {
		return submitBatch(requestContentForLLM, cfg, systemPrompt, mode, newClient)
	}

	result, err := callLLM(requestContentForLLM, flags.NoProgress, cfg, systemPrompt, newClient)
	if err != nil {
		return err
	}

	printCallResult(result)

	session := logic.Session{
		Linked:   requestContentForLLM.Linked,
		Mode:     mode,
		Response: result.Text,
	}

	if err := storeBuffer(session); err != nil {
		return err
	}

	return deliverResponse(session)
}

func indexSettingOrNone(setting *config.IndexSetting) config.IndexSetting {
	if setting != nil {
		return *setting
	}
	return config.IndexNone
}
