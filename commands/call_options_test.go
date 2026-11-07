package commands

import (
	"testing"

	"yact/config"
)

func TestValidateStepsOptions(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		mode     config.OutputMode
		prompt   string
		wantErr  bool
	}{
		{name: "empty selector accepts plan mode and prompt", selector: "", mode: config.OutputWritePlan, prompt: "x"},
		{name: "empty selector accepts apply mode with no prompt", selector: "", mode: config.OutputApply, prompt: ""},
		{name: "selector with plan mode fails", selector: "1", mode: config.OutputWritePlan, wantErr: true},
		{name: "selector with prompt fails", selector: "1", mode: config.OutputApply, prompt: "fix it", wantErr: true},
		{name: "list selector passes", selector: "1,3,5-6", mode: config.OutputApply},
		{name: "next passes", selector: "next", mode: config.OutputApply},
		{name: "invalid selector fails", selector: "abc", mode: config.OutputApply, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStepsOptions(tc.selector, tc.mode, tc.prompt)
			if tc.wantErr && err == nil {
				t.Fatal("validateStepsOptions() error = nil, want error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateStepsOptions() unexpected error: %v", err)
			}
		})
	}
}

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
