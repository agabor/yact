package commands

import (
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
