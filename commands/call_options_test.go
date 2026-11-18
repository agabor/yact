package commands

import (
	"testing"

	"yact/config"
)

func TestApplyModelOverride(t *testing.T) {
	tests := []struct {
		name     string
		override string
		want     string
	}{
		{name: "empty override keeps the configured model", override: "", want: config.ClaudeModel},
		{name: "non-empty override replaces the model", override: "opus", want: "opus"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			applyModelOverride(cfg, tc.override)
			if cfg.ClaudeModel != tc.want {
				t.Errorf("applyModelOverride(%q) ClaudeModel = %q, want %q", tc.override, cfg.ClaudeModel, tc.want)
			}
		})
	}
}
