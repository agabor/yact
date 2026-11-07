package commands

import (
	"fmt"

	"yact/config"
	"yact/index"
	"yact/logic"
)

func applyModelOverride(cfg *config.Config, modelOverride string) {
	if modelOverride != "" {
		cfg.ClaudeModel = modelOverride
	}
}

func applyIndexOverride(setting *config.IndexSetting) error {
	if setting == nil || *setting == config.IndexNone {
		return logic.RemoveIndexFile()
	}
	return index.WriteIndex(setting.Visibility())
}

func validateStepsOptions(stepsSelector string, mode config.OutputMode, prompt string) error {
	if stepsSelector == "" {
		return nil
	}
	if mode == config.OutputWritePlan {
		return fmt.Errorf("--steps cannot be combined with --output plan")
	}
	if prompt != "" {
		return fmt.Errorf("--steps cannot be combined with a prompt argument")
	}
	_, err := logic.ParseStepSelector(stepsSelector)
	return err
}
