package logic

import (
	"testing"

	"yact/config"
)

func TestIsInIgnoredFolder(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		relativePath string
		want         bool
	}{
		{name: "root folder ignores files inside it", pattern: "vendor", relativePath: "vendor/x.go", want: true},
		{name: "root folder does not ignore a file with the same prefix", pattern: "vendor", relativePath: "vendor.go", want: false},
		{name: "nested path ignores files inside it", pattern: "foo/bar", relativePath: "foo/bar/x.go", want: true},
		{name: "nested path is anchored at the root", pattern: "foo/bar", relativePath: "bar/foo/bar/x.go", want: false},
		{name: "single star matches one folder", pattern: "foo/*/bar", relativePath: "foo/a/bar/x.go", want: true},
		{name: "single star does not match two folders", pattern: "foo/*/bar", relativePath: "foo/a/b/bar/x.go", want: false},
		{name: "double star matches zero folders", pattern: "foo/**/bar", relativePath: "foo/bar/x.go", want: true},
		{name: "double star matches several folders", pattern: "foo/**/bar", relativePath: "foo/a/b/bar/x.go", want: true},
		{name: "file named like the folder is not ignored", pattern: "foo/bar", relativePath: "foo/bar", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pattern, err := config.ParseIgnorePattern(tc.pattern)
			if err != nil {
				t.Fatalf("ParseIgnorePattern(%q) unexpected error: %v", tc.pattern, err)
			}
			got := IsInIgnoredFolder(tc.relativePath, []config.IgnorePattern{pattern})
			if got != tc.want {
				t.Errorf("IsInIgnoredFolder(%q, %q) = %v, want %v", tc.relativePath, tc.pattern, got, tc.want)
			}
		})
	}
}
