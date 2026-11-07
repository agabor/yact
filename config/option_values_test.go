package config

import "testing"

func TestParseModelName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    ModelName
		wantErr bool
	}{
		{name: "haiku by name", raw: "haiku", want: ModelHaiku},
		{name: "sonnet by name", raw: "sonnet", want: ModelSonnet},
		{name: "opus by name", raw: "opus", want: ModelOpus},
		{name: "fable by name", raw: "fable", want: ModelFable},
		{name: "1 is haiku", raw: "1", want: ModelHaiku},
		{name: "2 is sonnet", raw: "2", want: ModelSonnet},
		{name: "3 is opus", raw: "3", want: ModelOpus},
		{name: "4 is fable", raw: "4", want: ModelFable},
		{name: "mixed case name", raw: "Opus", want: ModelOpus},
		{name: "number above range", raw: "5", wantErr: true},
		{name: "unknown name", raw: "claude", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseModelName(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseModelName(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseModelName(%q) unexpected error: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("ParseModelName(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseIndexSetting(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    IndexSetting
		wantErr bool
	}{
		{name: "none by name", raw: "none", want: IndexNone},
		{name: "public by name", raw: "public", want: IndexPublic},
		{name: "all is private", raw: "all", want: IndexPrivate},
		{name: "0 is none", raw: "0", want: IndexNone},
		{name: "1 is public", raw: "1", want: IndexPublic},
		{name: "2 is private", raw: "2", want: IndexPrivate},
		{name: "mixed case name", raw: "PUBLIC", want: IndexPublic},
		{name: "number above range", raw: "3", wantErr: true},
		{name: "negative number", raw: "-1", wantErr: true},
		{name: "unknown name", raw: "abc", wantErr: true},
		{name: "empty input", raw: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseIndexSetting(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseIndexSetting(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseIndexSetting(%q) unexpected error: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("ParseIndexSetting(%q) = %d, want %d", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseOutputName(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    OutputName
		wantErr bool
	}{
		{name: "apply", raw: "apply", want: OutputNameApply},
		{name: "print", raw: "print", want: OutputNamePrint},
		{name: "plan", raw: "plan", want: OutputNamePlan},
		{name: "mixed case", raw: "Plan", want: OutputNamePlan},
		{name: "unknown value", raw: "maybe", wantErr: true},
		{name: "empty input", raw: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseOutputName(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseOutputName(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseOutputName(%q) unexpected error: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("ParseOutputName(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestIndexSettingVisibility(t *testing.T) {
	tests := []struct {
		name    string
		setting IndexSetting
		want    string
	}{
		{name: "none has no visibility", setting: IndexNone, want: ""},
		{name: "public", setting: IndexPublic, want: "public"},
		{name: "all is private", setting: IndexPrivate, want: "private"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.setting.Visibility(); got != tc.want {
				t.Errorf("Visibility() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIndexSettingName(t *testing.T) {
	tests := []struct {
		name    string
		setting IndexSetting
		want    string
	}{
		{name: "none", setting: IndexNone, want: "none"},
		{name: "public", setting: IndexPublic, want: "public"},
		{name: "private is named all", setting: IndexPrivate, want: "all"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.setting.Name(); got != tc.want {
				t.Errorf("Name() = %q, want %q", got, tc.want)
			}
		})
	}
}
