package commands

import (
	"testing"

	"yact/logic"
)

func TestFormatPlanChanges(t *testing.T) {
	tests := []struct {
		name    string
		changes logic.PlanChanges
		want    string
	}{
		{
			name:    "unchanged goal with no step changes",
			changes: logic.PlanChanges{Goal: logic.GoalUnchanged},
			want:    "Goal: unchanged\nSteps new: 0\nSteps changed: 0\nSteps removed: 0",
		},
		{
			name:    "changed goal",
			changes: logic.PlanChanges{Goal: logic.GoalChanged},
			want:    "Goal: changed\nSteps new: 0\nSteps changed: 0\nSteps removed: 0",
		},
		{
			name:    "new goal",
			changes: logic.PlanChanges{Goal: logic.GoalNew},
			want:    "Goal: new\nSteps new: 0\nSteps changed: 0\nSteps removed: 0",
		},
		{
			name:    "non-zero step counts",
			changes: logic.PlanChanges{Goal: logic.GoalUnchanged, NewSteps: 2, ChangedSteps: 1, RemovedSteps: 3},
			want:    "Goal: unchanged\nSteps new: 2\nSteps changed: 1\nSteps removed: 3",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatPlanChanges(tc.changes); got != tc.want {
				t.Errorf("formatPlanChanges() = %q, want %q", got, tc.want)
			}
		})
	}
}
