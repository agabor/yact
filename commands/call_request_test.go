package commands

import "testing"

func TestShouldShowProgress(t *testing.T) {
	tests := []struct {
		name             string
		noProgress       bool
		stdoutIsTerminal bool
		want             bool
	}{
		{name: "terminal with progress enabled shows progress", noProgress: false, stdoutIsTerminal: true, want: true},
		{name: "terminal with no-progress hides progress", noProgress: true, stdoutIsTerminal: true, want: false},
		{name: "piped output with progress enabled hides progress", noProgress: false, stdoutIsTerminal: false, want: false},
		{name: "piped output with no-progress hides progress", noProgress: true, stdoutIsTerminal: false, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldShowProgress(tc.noProgress, tc.stdoutIsTerminal); got != tc.want {
				t.Errorf("shouldShowProgress(%v, %v) = %v, want %v", tc.noProgress, tc.stdoutIsTerminal, got, tc.want)
			}
		})
	}
}
