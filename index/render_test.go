package index

import (
	"strings"
	"testing"
)

func TestRenderFileSection(t *testing.T) {
	signatures := []IndexSignature{
		{Text: "func A()", Depth: 0},
		{Text: "func (t *T) B()", Depth: 1},
		{Text: "func C()", Depth: 2},
	}
	want := "pkg/a.go\n  func A()\n    func (t *T) B()\n      func C()\n"
	if got := renderFileSection("pkg/a.go", signatures); got != want {
		t.Errorf("renderFileSection() = %q, want %q", got, want)
	}
}

func TestRenderFileSectionWithoutSignatures(t *testing.T) {
	want := "empty.go\n"
	if got := renderFileSection("empty.go", nil); got != want {
		t.Errorf("renderFileSection() = %q, want %q", got, want)
	}
}

func TestRenderIndexHeaderForVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		wantScope  string
		absent     string
	}{
		{name: "public scope", visibility: "public", wantScope: publicIndexScope, absent: privateIndexScope},
		{name: "private scope", visibility: "private", wantScope: privateIndexScope, absent: publicIndexScope},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := &IndexNode{Name: ".", Type: "dir"}
			got := renderIndex(root, tc.visibility)
			if !strings.Contains(got, tc.wantScope) {
				t.Errorf("renderIndex(%q) does not contain %q", tc.visibility, tc.wantScope)
			}
			if strings.Contains(got, tc.absent) {
				t.Errorf("renderIndex(%q) contains %q", tc.visibility, tc.absent)
			}
		})
	}
}

func TestRenderIndexOrdersSectionsByPath(t *testing.T) {
	root := &IndexNode{
		Name: ".",
		Type: "dir",
		Children: []*IndexNode{
			{
				Name: "pkg",
				Type: "dir",
				Children: []*IndexNode{
					{
						Name:      "a.go",
						Type:      "file",
						Functions: []IndexSignature{{Text: "func A()", Depth: 0}},
					},
				},
			},
			{
				Name:      "b.go",
				Type:      "file",
				Functions: []IndexSignature{{Text: "func B()", Depth: 0}},
			},
		},
	}
	got := renderIndex(root, "public")
	if !strings.HasPrefix(got, renderIndexHeader("public")+"\n") {
		t.Errorf("renderIndex() does not start with the index header")
	}
	wantSections := "pkg/a.go\n  func A()\n\nb.go\n  func B()\n"
	if !strings.HasSuffix(got, wantSections) {
		t.Errorf("renderIndex() = %q, want sections ending with %q", got, wantSections)
	}
}