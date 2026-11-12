package logic

import (
	"reflect"
	"testing"
)

func TestParseAssumptions(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		want       []Assumption
	}{
		{
			name:       "ticked, unticked and X boxes with ids",
			promptText: planText("## Assumptions", "- [ ] **A1.** Use SQLite", "- [x] **A2.** Use Postgres", "- [X] **A3.** Keep cache"),
			want: []Assumption{
				{ID: "A1", Text: "**A1.** Use SQLite"},
				{ID: "A2", Text: "**A2.** Use Postgres", Checked: true},
				{ID: "A3", Text: "**A3.** Keep cache", Checked: true},
			},
		},
		{
			name:       "entry without box",
			promptText: planText("## Assumptions", "- **A1.** Use SQLite"),
			want:       []Assumption{{ID: "A1", Text: "**A1.** Use SQLite"}},
		},
		{
			name:       "entry without id",
			promptText: planText("## Assumptions", "- [ ] Users are logged in"),
			want:       []Assumption{{Text: "Users are logged in"}},
		},
		{
			name:       "answer on one line",
			promptText: planText("## Assumptions", "- [ ] **A1.** Cache", "  Answer: Use Redis"),
			want:       []Assumption{{ID: "A1", Text: "**A1.** Cache", Answer: "Use Redis"}},
		},
		{
			name:       "answer on several lines",
			promptText: planText("## Assumptions", "- [ ] **A1.** Cache", "  Answer: Use Redis", "  and keep it local."),
			want:       []Assumption{{ID: "A1", Text: "**A1.** Cache", Answer: "Use Redis\nand keep it local."}},
		},
		{
			name:       "empty answer counts as no answer",
			promptText: planText("## Assumptions", "- [ ] **A1.** Cache", "  Answer:"),
			want:       []Assumption{{ID: "A1", Text: "**A1.** Cache"}},
		},
		{
			name:       "section ends at a heading",
			promptText: planText("## Assumptions", "- [ ] **A1.** One", "", "## Steps", "- [ ] **A2.** Outside"),
			want:       []Assumption{{ID: "A1", Text: "**A1.** One"}},
		},
		{
			name:       "section ends at a separator",
			promptText: planText("## Assumptions", "- [ ] **A1.** One", "", "-----", "- [ ] **A2.** Outside"),
			want:       []Assumption{{ID: "A1", Text: "**A1.** One"}},
		},
		{
			name:       "section ends at the end of the text",
			promptText: planText("## Assumptions", "- [ ] **A1.** One"),
			want:       []Assumption{{ID: "A1", Text: "**A1.** One"}},
		},
		{
			name:       "no section gives nil",
			promptText: planText("# Goal", "- [ ] **A1.** Not a section"),
			want:       nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseAssumptions(tc.promptText)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseAssumptions() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestAssumptionIsUnanswered(t *testing.T) {
	tests := []struct {
		name       string
		assumption Assumption
		want       bool
	}{
		{name: "ticked", assumption: Assumption{ID: "A1", Checked: true}, want: false},
		{name: "unticked and unanswered", assumption: Assumption{ID: "A1"}, want: true},
		{name: "unticked with answer", assumption: Assumption{ID: "A1", Answer: "Use Redis"}, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.assumption.IsUnanswered(); got != tc.want {
				t.Errorf("IsUnanswered() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAssumptionHasUnappliedAnswer(t *testing.T) {
	tests := []struct {
		name       string
		assumption Assumption
		want       bool
	}{
		{name: "ticked", assumption: Assumption{ID: "A1", Checked: true}, want: false},
		{name: "unticked and unanswered", assumption: Assumption{ID: "A1"}, want: false},
		{name: "unticked with answer", assumption: Assumption{ID: "A1", Answer: "Use Redis"}, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.assumption.HasUnappliedAnswer(); got != tc.want {
				t.Errorf("HasUnappliedAnswer() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCarryAssumptionTicks(t *testing.T) {
	tests := []struct {
		name     string
		previous string
		next     string
		want     string
	}{
		{
			name:     "same text keeps the tick",
			previous: planText("## Assumptions", "- [x] **A1.** Cache"),
			next:     planText("## Assumptions", "- [ ] **A1.** Cache"),
			want:     planText("## Assumptions", "- [x] **A1.** Cache"),
		},
		{
			name:     "reworded text is unticked",
			previous: planText("## Assumptions", "- [x] **A1.** Cache"),
			next:     planText("## Assumptions", "- [ ] **A1.** Cache redis"),
			want:     planText("## Assumptions", "- [ ] **A1.** Cache redis"),
		},
		{
			name:     "tick written by the model is reset",
			previous: planText("## Assumptions", "- [ ] **A1.** Cache"),
			next:     planText("## Assumptions", "- [x] **A1.** Cache"),
			want:     planText("## Assumptions", "- [ ] **A1.** Cache"),
		},
		{
			name:     "new entry is unticked",
			previous: planText("## Assumptions", "- [x] **A1.** Cache"),
			next:     planText("## Assumptions", "- [x] **A1.** Cache", "- [x] **A2.** New"),
			want:     planText("## Assumptions", "- [x] **A1.** Cache", "- [ ] **A2.** New"),
		},
		{
			name:     "changed id is unticked",
			previous: planText("## Assumptions", "- [x] **A1.** Cache"),
			next:     planText("## Assumptions", "- [x] **A2.** Cache"),
			want:     planText("## Assumptions", "- [ ] **A2.** Cache"),
		},
		{
			name:     "no previous plan leaves entries unticked",
			previous: "",
			next:     planText("## Assumptions", "- [x] **A1.** Cache"),
			want:     planText("## Assumptions", "- [ ] **A1.** Cache"),
		},
		{
			name:     "entry without a box is left as it is",
			previous: planText("## Assumptions", "- [x] **A1.** Cache"),
			next:     planText("## Assumptions", "- **A1.** Cache"),
			want:     planText("## Assumptions", "- **A1.** Cache"),
		},
		{
			name:     "CRLF line endings are kept",
			previous: "## Assumptions\r\n- [x] **A1.** Cache\r\n",
			next:     "## Assumptions\r\n- [ ] **A1.** Cache\r\n",
			want:     "## Assumptions\r\n- [x] **A1.** Cache\r\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CarryAssumptionTicks(tc.previous, tc.next); got != tc.want {
				t.Errorf("CarryAssumptionTicks() = %q, want %q", got, tc.want)
			}
		})
	}
}
