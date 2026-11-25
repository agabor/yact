package logic

import "testing"

func TestComparePlans(t *testing.T) {
	basePlan := planText(
		"# Goal",
		"Ship it",
		"",
		"## Steps",
		"### 1. Write parser",
		"Parse input.",
		"",
		"### 2. Write output",
		"Print result.",
	)
	tests := []struct {
		name         string
		previousPlan string
		newPlan      string
		want         PlanChanges
	}{
		{
			name:         "identical plans have no changes",
			previousPlan: basePlan,
			newPlan:      basePlan,
			want:         PlanChanges{Goal: GoalUnchanged},
		},
		{
			name:         "reworded goal is changed",
			previousPlan: basePlan,
			newPlan: planText(
				"# Goal",
				"Ship it faster",
				"",
				"## Steps",
				"### 1. Write parser",
				"Parse input.",
				"",
				"### 2. Write output",
				"Print result.",
			),
			want: PlanChanges{Goal: GoalChanged},
		},
		{
			name:         "edited step body is one changed step",
			previousPlan: basePlan,
			newPlan: planText(
				"# Goal",
				"Ship it",
				"",
				"## Steps",
				"### 1. Write parser",
				"Parse input carefully.",
				"",
				"### 2. Write output",
				"Print result.",
			),
			want: PlanChanges{Goal: GoalUnchanged, ChangedSteps: 1},
		},
		{
			name:         "added step is one new step",
			previousPlan: basePlan,
			newPlan: planText(
				"# Goal",
				"Ship it",
				"",
				"## Steps",
				"### 1. Write parser",
				"Parse input.",
				"",
				"### 2. Write output",
				"Print result.",
				"",
				"### 3. Write tests",
				"Test it.",
			),
			want: PlanChanges{Goal: GoalUnchanged, NewSteps: 1},
		},
		{
			name:         "removed step is one removed step",
			previousPlan: basePlan,
			newPlan: planText(
				"# Goal",
				"Ship it",
				"",
				"## Steps",
				"### 1. Write parser",
				"Parse input.",
			),
			want: PlanChanges{Goal: GoalUnchanged, RemovedSteps: 1},
		},
		{
			name:         "renumbered step with same title and body is not a change",
			previousPlan: basePlan,
			newPlan: planText(
				"# Goal",
				"Ship it",
				"",
				"## Steps",
				"### 2. Write parser",
				"Parse input.",
				"",
				"### 1. Write output",
				"Print result.",
			),
			want: PlanChanges{Goal: GoalUnchanged},
		},
		{
			name:         "empty previous plan has new goal and all steps new",
			previousPlan: "",
			newPlan:      basePlan,
			want:         PlanChanges{Goal: GoalNew, NewSteps: 2},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ComparePlans(tc.previousPlan, tc.newPlan); got != tc.want {
				t.Errorf("ComparePlans() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
