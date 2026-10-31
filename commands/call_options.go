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

func applyMaxInputLinesOverride(cfg *config.Config, maxInputLinesOverride int) {
	if maxInputLinesOverride > 0 {
		cfg.MaxInputLines = maxInputLinesOverride
	}
}

func indexLevelToVisibility(level int) (string, error) {
	switch level {
	case 1:
		return "public", nil
	case 2:
		return "private", nil
	}
	return "", fmt.Errorf("invalid index level %d, expected 1 or 2", level)
}

func applyIndexOverride(flags config.Flaglist) error {
	if !flags.HasIndex() {
		return logic.RemoveIndexFile()
	}
	visibility, err := indexLevelToVisibility(flags.IndexLevel)
	if err != nil {
		return err
	}
	return index.WriteIndex(visibility)
}

func validateStepsOptions(stepsSelector string, writePlan bool, prompt string) error {
	if stepsSelector == "" {
		return nil
	}
	if writePlan {
		return fmt.Errorf("--steps and --plan cannot be combined")
	}
	if prompt != "" {
		return fmt.Errorf("--steps cannot be combined with a prompt argument")
	}
	_, err := logic.ParseStepSelector(stepsSelector)
	return err
}