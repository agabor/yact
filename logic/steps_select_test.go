package logic

import (
	"reflect"
	"testing"
)

func TestParseStepSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		want     StepSelector
		wantErr  bool
	}{
		{name: "whole plan", selector: "0", want: StepSelector{Kind: StepSelectorWholePlan}},
		{name: "next", selector: "next", want: StepSelector{Kind: StepSelectorNext}},
		{name: "next in any case with spaces", selector: " NEXT ", want: StepSelector{Kind: StepSelectorNext}},
		{name: "single number", selector: "3", want: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 3, End: 3}}}},
		{name: "range", selector: "2-4", want: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 2, End: 4}}}},
		{name: "comma list of numbers and ranges", selector: "1, 3-5", want: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 1, End: 1}, {Start: 3, End: 5}}}},
		{name: "empty selector", selector: "", wantErr: true},
		{name: "empty item", selector: "1,,2", wantErr: true},
		{name: "next combined with other items", selector: "next, 2", wantErr: true},
		{name: "reversed range", selector: "5-2", wantErr: true},
		{name: "zero step number", selector: "0,1", wantErr: true},
		{name: "non-numeric item", selector: "abc", wantErr: true},
		{name: "range with too many parts", selector: "1-2-3", wantErr: true},
		{name: "number too large", selector: "99999999999999999999", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseStepSelector(tc.selector)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseStepSelector(%q) error = nil, want error", tc.selector)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStepSelector(%q) unexpected error: %v", tc.selector, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseStepSelector(%q) = %+v, want %+v", tc.selector, got, tc.want)
			}
		})
	}
}

func TestSelectSteps(t *testing.T) {
	mixed := ParsePlan(planText("### [x] 1. A", "a", "### [ ] 2. B", "b", "### [x] 3. C", "c"))
	first := mixed.Steps[0]
	second := mixed.Steps[1]
	duplicates := ParsePlan(planText("### [ ] 1. A", "### [ ] 1. B"))
	noSteps := ParsePlan(planText("just prose"))
	allDone := ParsePlan(planText("### [x] 1. A"))
	tests := []struct {
		name     string
		plan     ParsedPlan
		selector StepSelector
		want     StepSelection
		wantErr  bool
	}{
		{
			name:     "whole plan is not validated",
			plan:     duplicates,
			selector: StepSelector{Kind: StepSelectorWholePlan},
			want:     StepSelection{},
		},
		{
			name:     "next returns the first unchecked step",
			plan:     mixed,
			selector: StepSelector{Kind: StepSelectorNext},
			want:     StepSelection{Steps: []PlanStep{second}},
		},
		{
			name:     "next reports all done when every step is checked",
			plan:     allDone,
			selector: StepSelector{Kind: StepSelectorNext},
			want:     StepSelection{AllDone: true},
		},
		{
			name:     "next on a plan without steps is an error",
			plan:     noSteps,
			selector: StepSelector{Kind: StepSelectorNext},
			wantErr:  true,
		},
		{
			name:     "range matching no step is an error",
			plan:     mixed,
			selector: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 5, End: 6}}},
			wantErr:  true,
		},
		{
			name:     "duplicate numbers are an error",
			plan:     duplicates,
			selector: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 1, End: 1}}},
			wantErr:  true,
		},
		{
			name:     "plan without steps is an error",
			plan:     noSteps,
			selector: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 1, End: 1}}},
			wantErr:  true,
		},
		{
			name:     "already checked step adds a warning",
			plan:     mixed,
			selector: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 1, End: 2}}},
			want: StepSelection{
				Steps:    []PlanStep{first, second},
				Warnings: []string{"step 1 is already done"},
			},
		},
		{
			name:     "single unchecked step has no warning",
			plan:     mixed,
			selector: StepSelector{Kind: StepSelectorRanges, Ranges: []StepRange{{Start: 2, End: 2}}},
			want:     StepSelection{Steps: []PlanStep{second}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectSteps(tc.plan, tc.selector)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("SelectSteps() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("SelectSteps() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SelectSteps() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestFormatNumbers(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    string
	}{
		{name: "empty slice", numbers: []int{}, want: ""},
		{name: "single number", numbers: []int{7}, want: "7"},
		{name: "multiple numbers", numbers: []int{1, 3, 5}, want: "1, 3, 5"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatNumbers(tc.numbers); got != tc.want {
				t.Errorf("FormatNumbers(%v) = %q, want %q", tc.numbers, got, tc.want)
			}
		})
	}
}