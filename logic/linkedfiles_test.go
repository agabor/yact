package logic

import (
	"testing"
	"yact/projectpath"
)

func TestLinkedFilesAllows(t *testing.T) {
	linked := LinkedFiles{"logic/links.go"}
	tests := []struct {
		name string
		path projectpath.Path
		want bool
	}{
		{"linked path", "logic/links.go", true},
		{"unlinked path", "logic/codefile.go", false},
		{"same name in another folder", "api/links.go", false},
		{"feedback file", FeedbackFilePath(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := linked.Allows(tt.path); got != tt.want {
				t.Errorf("LinkedFiles%v.Allows(%q) = %v, want %v", linked, tt.path, got, tt.want)
			}
		})
	}
}
