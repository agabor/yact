package commands

import (
	"fmt"
	"reflect"
	"testing"
)

func longWarningText(path string, lines int) string {
	return fmt.Sprintf("%s has %d lines, more than %d", path, lines, longFileLineThreshold)
}

func TestLongFileWarnings(t *testing.T) {
	tests := []struct {
		name   string
		counts []fileLineCount
		want   []string
	}{
		{
			name:   "empty counts return no warnings",
			counts: nil,
			want:   nil,
		},
		{
			name:   "exactly the threshold returns no warning",
			counts: []fileLineCount{{Path: "a.go", Lines: longFileLineThreshold}},
			want:   nil,
		},
		{
			name:   "one line over the threshold returns one warning",
			counts: []fileLineCount{{Path: "a.go", Lines: longFileLineThreshold + 1}},
			want:   []string{longWarningText("a.go", longFileLineThreshold+1)},
		},
		{
			name: "several long files return one warning each",
			counts: []fileLineCount{
				{Path: "a.go", Lines: longFileLineThreshold + 1},
				{Path: "small.go", Lines: 10},
				{Path: "b.go", Lines: longFileLineThreshold + 50},
			},
			want: []string{
				longWarningText("a.go", longFileLineThreshold+1),
				longWarningText("b.go", longFileLineThreshold+50),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := longFileWarnings(tc.counts)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("longFileWarnings() = %q, want %q", got, tc.want)
			}
		})
	}
}