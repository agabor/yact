package commands

import (
	"testing"

	"yact/config"
	"yact/projectpath"
)

func TestSkipsIgnoredFolder(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		want     bool
	}{
		{
			name:     "file inside a root folder is skipped",
			patterns: []string{"vendor"},
			path:     "vendor/x.go",
			want:     true,
		},
		{
			name:     "file whose name only shares the prefix is not skipped",
			patterns: []string{"vendor"},
			path:     "vendor.go",
			want:     false,
		},
		{
			name:     "nested pattern skips file below it",
			patterns: []string{"foo/bar"},
			path:     "foo/bar/x.go",
			want:     true,
		},
		{
			name:     "pattern is anchored at the project root",
			patterns: []string{"foo/bar"},
			path:     "bar/foo/bar/x.go",
			want:     false,
		},
		{
			name:     "empty ignore list skips nothing",
			patterns: []string{},
			path:     "vendor/x.go",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ignoredFolders []config.IgnorePattern
			for _, raw := range tt.patterns {
				pattern, err := config.ParseIgnorePattern(raw)
				if err != nil {
					t.Fatalf("ParseIgnorePattern(%q) failed: %v", raw, err)
				}
				ignoredFolders = append(ignoredFolders, pattern)
			}

			filePath, err := projectpath.Parse(tt.path)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.path, err)
			}

			if got := skipsIgnoredFolder(filePath, ignoredFolders); got != tt.want {
				t.Errorf("skipsIgnoredFolder(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
