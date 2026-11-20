package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func chdirToTempProject(t *testing.T) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restoring working directory: %v", err)
		}
	})
}

func TestParseExtensions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{name: "trims lines", content: "  go \npy\n", want: []string{"go", "py"}},
		{name: "skips blank lines", content: "go\n\n   \nrs", want: []string{"go", "rs"}},
		{name: "empty content returns defaults", content: "", want: DefaultExtensions()},
		{name: "only blank lines returns defaults", content: "\n  \n", want: DefaultExtensions()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseExtensions(tc.content)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseExtensions(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}

func TestLoadExtensionsDoesNotCreateMissingFile(t *testing.T) {
	chdirToTempProject(t)

	got, err := LoadExtensions()
	if err != nil {
		t.Fatalf("LoadExtensions() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, DefaultExtensions()) {
		t.Errorf("LoadExtensions() = %v, want %v", got, DefaultExtensions())
	}

	path := filepath.Clean(string(GetProjectExtensionsPath()))
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("LoadExtensions() created %s, want no file", path)
	}
}

func TestLoadIgnoredFoldersDoesNotCreateMissingFile(t *testing.T) {
	chdirToTempProject(t)

	folders, err := LoadIgnoredFolders()
	if err != nil {
		t.Fatalf("LoadIgnoredFolders() unexpected error: %v", err)
	}

	got := make([]string, 0, len(folders))
	for _, folder := range folders {
		got = append(got, folder.Text())
	}
	if !reflect.DeepEqual(got, DefaultIgnoredFolders()) {
		t.Errorf("LoadIgnoredFolders() = %v, want %v", got, DefaultIgnoredFolders())
	}

	path := filepath.Clean(string(GetProjectIgnorePath()))
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("LoadIgnoredFolders() created %s, want no file", path)
	}
}
