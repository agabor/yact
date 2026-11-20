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

func indexVisibility(setting config.IndexSetting) (index.Visibility, bool) {
	switch setting {
	case config.IndexPublic:
		return index.VisibilityPublic, true
	case config.IndexPrivate:
		return index.VisibilityPrivate, true
	}
	return 0, false
}

func applyIndexOverride(setting *config.IndexSetting) error {
	if setting != nil {
		if visibility, ok := indexVisibility(*setting); ok {
			return index.WriteIndex(visibility)
		}
	}
	return logic.RemoveIndexFile()
}
