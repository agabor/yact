package logic

import "testing"

func TestParseProjectPath(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    ProjectPath
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
			got, err := ParseProjectPath(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseProjectPath(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseProjectPath(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
