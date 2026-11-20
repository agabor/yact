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

func TestCountMissingContextEntries(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		want       int
	}{
		{name: "no section gives zero", promptText: planText("# Goal", "- a.go"), want: 0},
		{name: "two top-level items", promptText: planText("## Missing context", "- a.go", "- b.go"), want: 2},
		{name: "indented continuation is not counted", promptText: planText("## Missing context", "- a.go", "  needed for the cache", "  - nested"), want: 1},
		{name: "items below the next heading are not counted", promptText: planText("## Missing context", "- a.go", "", "## Steps", "- [ ] 1. Do"), want: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CountMissingContextEntries(tc.promptText); got != tc.want {
				t.Errorf("CountMissingContextEntries() = %d, want %d", got, tc.want)
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
		{name: "missing context section", promptText: planText("# Goal", "", "## Missing context", "- a.go"), want: PromptGateMissingContext},
		{name: "empty prompt is open", promptText: "", want: PromptGateOpen},
		{name: "unanswered assumption", promptText: planText("# Goal", "", "## Assumptions", "- [ ] **A1.** Cache"), want: PromptGateUnansweredAssumptions},
		{name: "answered but unticked assumption", promptText: planText("## Assumptions", "- [ ] **A1.** Cache", "  Answer: Use Redis"), want: PromptGateUnappliedAnswers},
		{name: "ticked assumption with answer is open", promptText: planText("## Assumptions", "- [x] **A1.** Cache", "  Answer: Use Redis"), want: PromptGateOpen},
		{name: "missing context with unanswered assumption", promptText: planText("## Missing context", "- a.go", "## Assumptions", "- [ ] **A1.** Cache"), want: PromptGateMissingContext},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecidePromptGate(tc.promptText); got != tc.want {
				t.Errorf("DecidePromptGate() = %v, want %v", got, tc.want)
			}
		})
	}
}
