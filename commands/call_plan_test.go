package commands

import (
	"strings"
	"testing"

	"yact/logic"
)

const samplePlan = "# Plan\n\n" +
	"### [ ] 1. First\nbody one\n\n" +
	"### [x] 2. Second\nbody two\n\n" +
	"### [ ] 3. Third\nbody three\n"

const allUncheckedPlan = "# Plan\n\n" +
	"### [ ] 1. First\nbody one\n\n" +
	"### [ ] 2. Second\nbody two\n"

const allCheckedPlan = "# Plan\n\n" +
	"### [x] 1. First\nbody one\n\n" +
	"### [x] 2. Second\nbody two\n"

const prosePlan = "# Notes\nno steps here\n"

const promptUncheckedSamplePlan = "# Plan\n\n" +
	"### [ ] 1. First\nbody one\n\n" +
	"### [ ] 3. Third\nbody three\n"

const promptStepOne = "# Plan\n\n### [ ] 1. First\nbody one\n\n"

const promptStepTwo = "# Plan\n\n### [x] 2. Second\nbody two\n\n"

func TestCombinePromptText(t *testing.T) {
	tests := []struct {
		name          string
		promptFile    string
		cliPrompt     string
		want          string
	}{
		{name: "both inputs empty returns empty", promptFile: "", cliPrompt: "", want: ""},
		{name: "file text only", promptFile: "  file text \n", cliPrompt: "", want: "file text"},
		{name: "CLI text only", promptFile: " \n", cliPrompt: "cli text", want: "cli text"},
		{name: "both joined with a blank line after trimming", promptFile: " file \n", cliPrompt: "  cli  ", want: "file\n\ncli"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := combinePromptText(tc.promptFile, tc.cliPrompt); got != tc.want {
				t.Errorf("combinePromptText(%q, %q) = %q, want %q", tc.promptFile, tc.cliPrompt, got, tc.want)
			}
		})
	}
}

func TestSelectPlanSteps(t *testing.T) {
	tests := []struct {
		name         string
		planText     string
		selector     string
		wantErr      bool
		wantAllDone  bool
		wantPrompt   string
		wantSteps    string
		wantWarnings string
	}{
		{
			name:       "no selector with all steps unchecked returns the full plan",
			planText:   allUncheckedPlan,
			selector:   "",
			wantPrompt: allUncheckedPlan,
			wantSteps:  "1, 2",
		},
		{
			name:       "no selector with some steps checked returns only unchecked steps",
			planText:   samplePlan,
			selector:   "",
			wantPrompt: promptUncheckedSamplePlan,
			wantSteps:  "1, 3",
		},
		{
			name:        "no selector with all steps checked sets AllDone",
			planText:    allCheckedPlan,
			selector:    "",
			wantAllDone: true,
		},
		{
			name:       "no selector with a plan of zero steps uses the full plan",
			planText:   prosePlan,
			selector:   "",
			wantPrompt: prosePlan,
		},
		{
			name:       "selector 0 returns the whole plan without steps",
			planText:   samplePlan,
			selector:   "0",
			wantPrompt: samplePlan,
		},
		{
			name:       "specific step selector returns only that step",
			planText:   samplePlan,
			selector:   "1",
			wantPrompt: promptStepOne,
			wantSteps:  "1",
		},
		{
			name:         "selecting a done step passes its warning through",
			planText:     samplePlan,
			selector:     "2",
			wantPrompt:   promptStepTwo,
			wantSteps:    "2",
			wantWarnings: "step 2 is already done",
		},
		{
			name:       "next selects the first unchecked step",
			planText:   samplePlan,
			selector:   "next",
			wantPrompt: promptStepOne,
			wantSteps:  "1",
		},
		{
			name:        "next on an all-checked plan sets AllDone",
			planText:    allCheckedPlan,
			selector:    "next",
			wantAllDone: true,
		},
		{
			name:     "invalid selector returns an error",
			planText: samplePlan,
			selector: "abc",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selection, err := selectPlanSteps(tc.planText, tc.selector)
			if tc.wantErr {
				if err == nil {
					t.Fatal("selectPlanSteps() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("selectPlanSteps() unexpected error: %v", err)
			}
			if selection.AllDone != tc.wantAllDone {
				t.Errorf("selectPlanSteps() AllDone = %t, want %t", selection.AllDone, tc.wantAllDone)
			}
			if selection.Prompt != tc.wantPrompt {
				t.Errorf("selectPlanSteps() Prompt = %q, want %q", selection.Prompt, tc.wantPrompt)
			}
			if got := logic.FormatNumbers(logic.StepNumbers(selection.Steps)); got != tc.wantSteps {
				t.Errorf("selectPlanSteps() steps = %q, want %q", got, tc.wantSteps)
			}
			if got := strings.Join(selection.Warnings, "; "); got != tc.wantWarnings {
				t.Errorf("selectPlanSteps() warnings = %q, want %q", got, tc.wantWarnings)
			}
		})
	}
}