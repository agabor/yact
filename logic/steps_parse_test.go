package logic

import (
	"reflect"
	"strings"
	"testing"
)

func planText(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

func TestParsePlan(t *testing.T) {
	tests := []struct {
		name string
		plan string
		want ParsedPlan
	}{
		{
			name: "preamble, checked, unchecked and X checkboxes",
			plan: planText("# Goal", "", "Intro text.", "", "## Steps", "", "### [ ] 1. First", "", "Body one.", "", "### [x] 2. Second", "", "Body two.", "", "### [X] 3. Third"),
			want: ParsedPlan{
				Preamble: "# Goal\n\nIntro text.\n\n## Steps\n\n",
				Steps: []PlanStep{
					{Number: 1, Title: "First", Checked: false, HeadingLine: 6, Text: "### [ ] 1. First\n\nBody one.\n\n"},
					{Number: 2, Title: "Second", Checked: true, HeadingLine: 10, Text: "### [x] 2. Second\n\nBody two.\n\n"},
					{Number: 3, Title: "Third", Checked: true, HeadingLine: 14, Text: "### [X] 3. Third\n"},
				},
			},
		},
		{
			name: "plain heading without checkbox",
			plan: planText("### 1. Plain", "Body"),
			want: ParsedPlan{
				Steps: []PlanStep{
					{Number: 1, Title: "Plain", Checked: false, HeadingLine: 0, Text: "### 1. Plain\nBody\n"},
				},
			},
		},
		{
			name: "level two heading ends the step",
			plan: planText("### [ ] 1. A", "body", "## Next", "more"),
			want: ParsedPlan{
				Steps: []PlanStep{
					{Number: 1, Title: "A", Checked: false, HeadingLine: 0, Text: "### [ ] 1. A\nbody\n"},
				},
			},
		},
		{
			name: "plan without steps keeps the whole plan as preamble",
			plan: planText("# Title", "text"),
			want: ParsedPlan{Preamble: "# Title\ntext\n"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParsePlan(tc.plan)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParsePlan() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParsePlanCRLF(t *testing.T) {
	plan := strings.ReplaceAll(planText("### [x] 1. Done", "body", "### [ ] 2. Open"), "\n", "\r\n")
	steps := ParsePlan(plan).Steps
	tests := []struct {
		name        string
		index       int
		wantTitle   string
		wantChecked bool
	}{
		{name: "checked step title has no carriage return", index: 0, wantTitle: "Done", wantChecked: true},
		{name: "unchecked step title has no carriage return", index: 1, wantTitle: "Open", wantChecked: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.index >= len(steps) {
				t.Fatalf("got %d steps, want at least %d", len(steps), tc.index+1)
			}
			step := steps[tc.index]
			if strings.ContainsAny(step.Title, "\r\n") {
				t.Errorf("Title = %q, want no line break characters", step.Title)
			}
			if step.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", step.Title, tc.wantTitle)
			}
			if step.Checked != tc.wantChecked {
				t.Errorf("Checked = %v, want %v", step.Checked, tc.wantChecked)
			}
		})
	}
}

func TestParseStepHeading(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantNumber  int
		wantTitle   string
		wantChecked bool
		wantHeading bool
	}{
		{name: "unchecked checkbox", line: "### [ ] 1. Write tests\n", wantNumber: 1, wantTitle: "Write tests", wantChecked: false, wantHeading: true},
		{name: "lowercase x checkbox", line: "### [x] 2. Done\n", wantNumber: 2, wantTitle: "Done", wantChecked: true, wantHeading: true},
		{name: "uppercase X checkbox", line: "### [X] 3. Also done\n", wantNumber: 3, wantTitle: "Also done", wantChecked: true, wantHeading: true},
		{name: "plain heading without checkbox", line: "### 4. Plain\n", wantNumber: 4, wantTitle: "Plain", wantChecked: false, wantHeading: true},
		{name: "level two heading is not a step", line: "## [ ] 1. x", wantNumber: 0, wantTitle: "", wantChecked: false, wantHeading: false},
		{name: "non-numeric step number is not a step", line: "### [ ] abc. x", wantNumber: 0, wantTitle: "", wantChecked: false, wantHeading: false},
		{name: "text before the marker is not a step", line: "text ### [ ] 1. x", wantNumber: 0, wantTitle: "", wantChecked: false, wantHeading: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			number, title, checked, isHeading := parseStepHeading(tc.line)
			if number != tc.wantNumber || title != tc.wantTitle || checked != tc.wantChecked || isHeading != tc.wantHeading {
				t.Errorf("parseStepHeading(%q) = (%d, %q, %v, %v), want (%d, %q, %v, %v)",
					tc.line, number, title, checked, isHeading,
					tc.wantNumber, tc.wantTitle, tc.wantChecked, tc.wantHeading)
			}
		})
	}
}

func TestValidateDetectsDuplicateNumbers(t *testing.T) {
	tests := []struct {
		name    string
		plan    ParsedPlan
		wantErr bool
	}{
		{name: "unique numbers", plan: ParsedPlan{Steps: []PlanStep{{Number: 1}, {Number: 2}}}},
		{name: "no steps", plan: ParsedPlan{}},
		{name: "duplicate number", plan: ParsedPlan{Steps: []PlanStep{{Number: 2}, {Number: 2}}}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.plan.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestAllCheckedAndUncheckedSteps(t *testing.T) {
	finished := PlanStep{Number: 1, Title: "finished", Checked: true}
	pending := PlanStep{Number: 2, Title: "pending"}
	tests := []struct {
		name        string
		plan        ParsedPlan
		wantChecked bool
		wantPending []PlanStep
	}{
		{name: "no steps", plan: ParsedPlan{}, wantChecked: false, wantPending: []PlanStep{}},
		{name: "all checked", plan: ParsedPlan{Steps: []PlanStep{finished}}, wantChecked: true, wantPending: []PlanStep{}},
		{name: "mixed", plan: ParsedPlan{Steps: []PlanStep{finished, pending}}, wantChecked: false, wantPending: []PlanStep{pending}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.plan.AllChecked(); got != tc.wantChecked {
				t.Errorf("AllChecked() = %v, want %v", got, tc.wantChecked)
			}
			if got := tc.plan.UncheckedSteps(); !reflect.DeepEqual(got, tc.wantPending) {
				t.Errorf("UncheckedSteps() = %+v, want %+v", got, tc.wantPending)
			}
		})
	}
}