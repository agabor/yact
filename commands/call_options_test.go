package commands

import (
	"testing"

	"yact/config"
)

func TestIndexLevelToVisibility(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		want    string
		wantErr bool
	}{
		{name: "level 1 is public", level: 1, want: "public"},
		{name: "level 2 is private", level: 2, want: "private"},
		{name: "level 0 is invalid", level: 0, wantErr: true},
		{name: "level 3 is invalid", level: 3, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := indexLevelToVisibility(tc.level)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("indexLevelToVisibility(%d) error = nil, want error", tc.level)
				}
				return
			}
			if err != nil {
				t.Fatalf("indexLevelToVisibility(%d) unexpected error: %v", tc.level, err)
			}
			if got != tc.want {
				t.Errorf("indexLevelToVisibility(%d) = %q, want %q", tc.level, got, tc.want)
			}
		})
	}
}

func TestValidateStepsOptions(t *testing.T) {
	tests := []struct {
		name      string
		selector  string
		writePlan bool
		prompt    string
		wantErr   bool
	}{
		{name: "empty selector accepts plan flag and prompt", selector: "", writePlan: true, prompt: "x"},
		{name: "empty selector accepts no options", selector: "", writePlan: false, prompt: ""},
		{name: "selector with plan flag fails", selector: "1", writePlan: true, wantErr: true},
		{name: "selector with prompt fails", selector: "1", writePlan: false, prompt: "fix it", wantErr: true},
		{name: "list selector passes", selector: "1,3,5-6"},
		{name: "next passes", selector: "next"},
		{name: "invalid selector fails", selector: "abc", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStepsOptions(tc.selector, tc.writePlan, tc.prompt)
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

func TestApplyMaxInputLinesOverride(t *testing.T) {
	tests := []struct {
		name     string
		override int
		want     int
	}{
		{name: "zero keeps the configured limit", override: 0, want: config.MaxInputLines},
		{name: "negative keeps the configured limit", override: -5, want: config.MaxInputLines},
		{name: "positive value replaces the limit", override: 500, want: 500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			applyMaxInputLinesOverride(cfg, tc.override)
			if cfg.MaxInputLines != tc.want {
				t.Errorf("applyMaxInputLinesOverride(%d) MaxInputLines = %d, want %d", tc.override, cfg.MaxInputLines, tc.want)
			}
		})
	}
}