package logic

import "testing"

func TestBuildStepsPrompt(t *testing.T) {
	plan := ParsePlan(planText("# Goal", "", "### [ ] 1. A", "a", "### [ ] 2. B", "b"))
	first := plan.Steps[0]
	second := plan.Steps[1]
	tests := []struct {
		name  string
		steps []PlanStep
		want  string
	}{
		{name: "no steps keeps only the preamble", steps: nil, want: "# Goal\n\n"},
		{name: "steps follow the preamble in order", steps: []PlanStep{first, second}, want: "# Goal\n\n### [ ] 1. A\na\n### [ ] 2. B\nb\n"},
		{name: "steps keep the given order", steps: []PlanStep{second, first}, want: "# Goal\n\n### [ ] 2. B\nb\n### [ ] 1. A\na\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildStepsPrompt(plan, tc.steps); got != tc.want {
				t.Errorf("BuildStepsPrompt() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMarkLineDone(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "unchecked heading", line: "### [ ] 1. A\n", want: "### [x] 1. A\n"},
		{name: "uppercase checked heading", line: "### [X] 1. A\n", want: "### [x] 1. A\n"},
		{name: "heading without checkbox", line: "### 1. A\n", want: "### [x] 1. A\n"},
		{name: "heading keeps CRLF line ending", line: "### [ ] 1. A\r\n", want: "### [x] 1. A\r\n"},
		{name: "non-heading line unchanged", line: "Some text\n", want: "Some text\n"},
		{name: "level two heading unchanged", line: "## [ ] 1. A\n", want: "## [ ] 1. A\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := markLineDone(tc.line); got != tc.want {
				t.Errorf("markLineDone(%q) = %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

func TestMarkStepsDone(t *testing.T) {
	plan := planText("# Goal", "### [ ] 1. A", "a", "### [ ] 2. B", "b", "### [x] 3. C")
	parsed := ParsePlan(plan)
	first := parsed.Steps[0]
	second := parsed.Steps[1]
	checked := parsed.Steps[2]
	outsideOfPlan := PlanStep{Number: 9, HeadingLine: 99}
	negativeLine := PlanStep{Number: 8, HeadingLine: -1}
	tests := []struct {
		name  string
		steps []PlanStep
		want  string
	}{
		{
			name:  "only the selected heading changes",
			steps: []PlanStep{first},
			want:  "# Goal\n### [x] 1. A\na\n### [ ] 2. B\nb\n### [x] 3. C\n",
		},
		{
			name:  "several selected headings change",
			steps: []PlanStep{first, second},
			want:  "# Goal\n### [x] 1. A\na\n### [x] 2. B\nb\n### [x] 3. C\n",
		},
		{
			name:  "no selected steps leaves the plan unchanged",
			steps: nil,
			want:  plan,
		},
		{
			name:  "already checked heading stays checked",
			steps: []PlanStep{checked},
			want:  plan,
		},
		{
			name:  "heading line outside the plan is ignored",
			steps: []PlanStep{outsideOfPlan},
			want:  plan,
		},
		{
			name:  "negative heading line is ignored",
			steps: []PlanStep{negativeLine},
			want:  plan,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := MarkStepsDone(plan, tc.steps); got != tc.want {
				t.Errorf("MarkStepsDone() = %q, want %q", got, tc.want)
			}
		})
	}
}