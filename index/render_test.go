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
		visibility Visibility
		wantScope  string
		absent     string
	}{
		{name: "public scope", visibility: VisibilityPublic, wantScope: publicIndexScope, absent: privateIndexScope},
		{name: "private scope", visibility: VisibilityPrivate, wantScope: privateIndexScope, absent: publicIndexScope},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := &indexDir{Name: "."}
			got := renderIndex(root, tc.visibility)
			if !strings.Contains(got, tc.wantScope) {
				t.Errorf("renderIndex(%v) does not contain %q", tc.visibility, tc.wantScope)
			}
			if strings.Contains(got, tc.absent) {
				t.Errorf("renderIndex(%v) contains %q", tc.visibility, tc.absent)
			}
		})
	}
}

func TestRenderIndexOrdersSectionsByPath(t *testing.T) {
	root := &indexDir{
		Name: ".",
		Children: []indexEntry{
			&indexDir{
				Name: "pkg",
				Children: []indexEntry{
					&indexFile{
						Name:      "a.go",
						Functions: []IndexSignature{{Text: "func A()", Depth: 0}},
					},
				},
			},
			&indexFile{
				Name:      "b.go",
				Functions: []IndexSignature{{Text: "func B()", Depth: 0}},
			},
		},
	}
	got := renderIndex(root, VisibilityPublic)
	if !strings.HasPrefix(got, renderIndexHeader(VisibilityPublic)+"\n") {
		t.Errorf("renderIndex() does not start with the index header")
	}
	wantSections := "pkg/a.go\n  func A()\n\nb.go\n  func B()\n"
	if !strings.HasSuffix(got, wantSections) {
		t.Errorf("renderIndex() = %q, want sections ending with %q", got, wantSections)
	}
}
