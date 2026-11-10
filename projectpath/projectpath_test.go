package projectpath

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Path
		wantErr bool
	}{
		{"forward slashes kept", "logic/links.go", "logic/links.go", false},
		{"backslashes become slashes", `logic\links.go`, "logic/links.go", false},
		{"leading dot slash removed", "./main.go", "main.go", false},
		{"dot dot inside project cleaned", "a/../b", "b", false},
		{"empty rejected", "", "", true},
		{"whitespace only rejected", "   ", "", true},
		{"absolute path rejected", "/etc/passwd", "", true},
		{"backslash absolute rejected", `\etc\passwd`, "", true},
		{"windows drive letter rejected", `C:\x.go`, "", true},
		{"forward slash drive letter rejected", "C:/x.go", "", true},
		{"UNC path rejected", `\\server\share`, "", true},
		{"parent directory rejected", "../x.go", "", true},
		{"parent directory after backslashes rejected", `..\x.go`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name    string
		base    Path
		element string
		want    Path
	}{
		{"plain join", ".yact", "temp", ".yact/temp"},
		{"base with trailing slash", ".yact/", "extensions.txt", ".yact/extensions.txt"},
		{"nested element", ".yact", "temp/buffer.txt", ".yact/temp/buffer.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Join(tt.base, tt.element)
			if got != tt.want {
				t.Errorf("Join(%q, %q) = %q, want %q", tt.base, tt.element, got, tt.want)
			}
			if strings.Contains(string(got), `\`) {
				t.Errorf("Join(%q, %q) = %q, must not contain a backslash", tt.base, tt.element, got)
			}
		})
	}
}
