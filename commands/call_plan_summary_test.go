package commands

import (
	"strings"
	"testing"
)

func summaryPlanText(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func TestFormatPlanSummary(t *testing.T) {
	tests := []struct {
		name string
		plan string
		want string
	}{
		{
			name: "plan without sections",
			plan: summaryPlanText("# Goal", "", "Do the work"),
			want: "Assumptions: 0\nMissing context: 0",
		},
		{
			name: "two assumptions and one missing context entry",
			plan: summaryPlanText("## Missing context", "- a.go", "", "## Assumptions", "- [ ] **A1.** Cache", "- [x] **A2.** Logs"),
			want: "Assumptions: 2\nMissing context: 1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatPlanSummary(tc.plan); got != tc.want {
				t.Errorf("formatPlanSummary() = %q, want %q", got, tc.want)
			}
		})
	}
}
