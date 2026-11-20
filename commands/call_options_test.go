package commands

import (
	"testing"

	"yact/config"
)

func TestApplyModelOverride(t *testing.T) {
	opus := config.ModelOpus
	tests := []struct {
		name     string
		override *config.ModelName
		want     config.ModelName
	}{
		{name: "nil override keeps the configured model", override: nil, want: config.DefaultConfig().ClaudeModel},
		{name: "non-nil override replaces the model", override: &opus, want: config.ModelOpus},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			applyModelOverride(cfg, tc.override)
			if cfg.ClaudeModel != tc.want {
				t.Errorf("applyModelOverride ClaudeModel = %q, want %q", cfg.ClaudeModel, tc.want)
			}
		})
	}
}
