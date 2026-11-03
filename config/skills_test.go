package config

import "testing"

func TestParseSkillName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid name", raw: "review", wantErr: false},
		{name: "empty name", raw: "", wantErr: true},
		{name: "contains slash", raw: "a/b", wantErr: true},
		{name: "contains backslash", raw: `a\b`, wantErr: true},
		{name: "parent directory", raw: "..", wantErr: true},
		{name: "parent directory inside name", raw: "a..b", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSkillName(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSkillName(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if err == nil && string(got) != tt.raw {
				t.Errorf("ParseSkillName(%q) = %q, want %q", tt.raw, got, tt.raw)
			}
		})
	}
}
