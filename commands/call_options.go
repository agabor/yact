package commands

import (
	"yact/config"
	"yact/index"
	"yact/logic"
)

func applyModelOverride(cfg *config.Config, modelOverride *config.ModelName) {
	if modelOverride != nil {
		cfg.ClaudeModel = *modelOverride
	}
}

func applyIndexOverride(setting *config.IndexSetting) error {
	if setting == nil || *setting == config.IndexNone {
		return logic.RemoveIndexFile()
	}
	return index.WriteIndex(setting.Visibility())
}
