package logic

import "testing"

func TestHasMissingContextSection(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		want       bool
	}{
		{name: "heading present", promptText: planText("# Goal", "", "## Missing context", "- a.go"), want: true},
		{name: "heading absent", promptText: planText("# Goal", "", "## Steps"), want: false},
		{name: "level three heading does not count", promptText: planText("### Missing context"), want: false},
		{name: "heading with suffix does not count", promptText: planText("## Missing context notes"), want: false},
		{name: "CRLF line ending", promptText: "## Missing context\r\n", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasMissingContextSection(tc.promptText); got != tc.want {
				t.Errorf("HasMissingContextSection() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecidePromptGate(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		want       PromptGate
	}{
		{name: "open plan with unticked steps", promptText: planText("### [ ] 1. A", "body"), want: PromptGateOpen},
		{name: "fully ticked plan", promptText: planText("### [x] 1. A", "body"), want: PromptGateNoStepsRemain},
		{name: "missing context section", promptText: planText("# Goal", "", "## Missing context", "- a.go"), want: PromptGateMissingContext},
		{name: "ticked plan with missing context wins", promptText: planText("### [x] 1. A", "", "## Missing context"), want: PromptGateMissingContext},
		{name: "empty prompt is open", promptText: "", want: PromptGateOpen},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecidePromptGate(tc.promptText); got != tc.want {
				t.Errorf("DecidePromptGate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPromptGateError(t *testing.T) {
	tests := []struct {
		name    string
		gate    PromptGate
		wantErr string
	}{
		{name: "open gate has no error", gate: PromptGateOpen, wantErr: ""},
		{name: "missing context error", gate: PromptGateMissingContext, wantErr: "Files needed as context are missing; run with --plan to update the plan"},
		{name: "no steps remain error", gate: PromptGateNoStepsRemain, wantErr: "no steps remain; run with --plan to update the plan"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := PromptGateError(tc.gate)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("PromptGateError() = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("PromptGateError() = %v, want %q", err, tc.wantErr)
			}
		})
	}
}