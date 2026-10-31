package commands

import "testing"

func TestParseIndexLevel(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int
		wantErr bool
	}{
		{name: "level 0", arg: "0", want: 0},
		{name: "level 1", arg: "1", want: 1},
		{name: "level 2", arg: "2", want: 2},
		{name: "number above range", arg: "3", wantErr: true},
		{name: "negative number", arg: "-1", wantErr: true},
		{name: "non-numeric input", arg: "abc", wantErr: true},
		{name: "empty input", arg: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseIndexLevel(tc.arg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseIndexLevel(%q) error = nil, want error", tc.arg)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseIndexLevel(%q) unexpected error: %v", tc.arg, err)
			}
			if got != tc.want {
				t.Errorf("parseIndexLevel(%q) = %d, want %d", tc.arg, got, tc.want)
			}
		})
	}
}