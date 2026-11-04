package config

import "testing"

func indexLevelPointer(level int) *int {
	return &level
}

func TestApplyDefaultsKeepsExplicitIndexZero(t *testing.T) {
	tests := []struct {
		name     string
		defaults []string
	}{
		{name: "public default", defaults: []string{"i1"}},
		{name: "private default", defaults: []string{"i2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := Flaglist{IndexLevel: indexLevelPointer(0)}
			flags.ApplyDefaults(tt.defaults, false)
			if flags.IndexLevel == nil || *flags.IndexLevel != 0 {
				t.Fatalf("IndexLevel = %v, want 0", flags.IndexLevel)
			}
			if flags.HasIndex() {
				t.Error("HasIndex() = true, want false")
			}
		})
	}
}

func TestApplyDefaultsUsesIndexWhenUnset(t *testing.T) {
	tests := []struct {
		name     string
		defaults []string
		want     int
	}{
		{name: "public default", defaults: []string{"i1"}, want: 1},
		{name: "private default", defaults: []string{"i2"}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flags Flaglist
			flags.ApplyDefaults(tt.defaults, false)
			if flags.IndexLevel == nil || *flags.IndexLevel != tt.want {
				t.Fatalf("IndexLevel = %v, want %d", flags.IndexLevel, tt.want)
			}
		})
	}
}

func TestApplyDefaultsKeepsExplicitIndexLevel(t *testing.T) {
	tests := []struct {
		name     string
		explicit int
		defaults []string
	}{
		{name: "public kept over private default", explicit: 1, defaults: []string{"i2"}},
		{name: "private kept over public default", explicit: 2, defaults: []string{"i1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := Flaglist{IndexLevel: indexLevelPointer(tt.explicit)}
			flags.ApplyDefaults(tt.defaults, false)
			if flags.IndexLevel == nil || *flags.IndexLevel != tt.explicit {
				t.Fatalf("IndexLevel = %v, want %d", flags.IndexLevel, tt.explicit)
			}
		})
	}
}
