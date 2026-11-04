package logic

import (
	"slices"
	"testing"
)

func TestParsePromptLinkTargetsDoubleBrackets(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{"double bracket", "see [[main.go]]", []string{"main.go"}},
		{"markdown link", "see [main.go](main.go)", []string{"main.go"}},
		{"no link", "plain text with main.go", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePromptLinkTargets(tt.prompt)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ParsePromptLinkTargets(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestParsePromptLinkTargetsMixed(t *testing.T) {
	got := ParsePromptLinkTargets("[[a.go]] and [b.go](b.go)")
	for _, want := range []string{"a.go", "b.go"} {
		if !slices.Contains(got, want) {
			t.Errorf("ParsePromptLinkTargets() = %v, missing %q", got, want)
		}
	}
}

func TestParseMarkdownLinkTargets(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   []string
	}{
		{"markdown link", "see [main.go](main.go)", []string{"main.go"}},
		{"double bracket excluded", "see [[main.go]]", nil},
		{"mixed", "[[a.go]] and [b.go](b.go)", []string{"b.go"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMarkdownLinkTargets(tt.prompt)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ParseMarkdownLinkTargets(%q) = %v, want %v", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestParseGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    GlobPattern
		wantErr bool
	}{
		{"glob", "logic/*.go", "logic/*.go", false},
		{"exact path", "logic/links.go", "logic/links.go", false},
		{"malformed", "logic/[", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGlobPattern(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseGlobPattern(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseGlobPattern(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestGlobPatternMatches(t *testing.T) {
	tests := []struct {
		name    string
		pattern GlobPattern
		path    string
		want    bool
	}{
		{"star matches file in folder", "logic/*.go", "logic/links.go", true},
		{"star does not match other extension", "logic/*.go", "logic/links.txt", false},
		{"star does not cross slash", "logic/*.go", "logic/sub/x.go", false},
		{"star does not match other folder", "logic/*.go", "api/claude.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.Matches(tt.path); got != tt.want {
				t.Errorf("GlobPattern(%q).Matches(%q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestSelectLinkedFiles(t *testing.T) {
	files := []string{"logic/links.go", "logic/links_test.go", "main.go"}

	tests := []struct {
		name       string
		exactPaths []string
		patterns   []GlobPattern
		want       []string
	}{
		{
			name:       "exact path selects file",
			exactPaths: []string{"main.go"},
			want:       []string{"main.go"},
		},
		{
			name:     "glob pattern selects matching files",
			patterns: []GlobPattern{"logic/*.go"},
			want:     []string{"logic/links.go", "logic/links_test.go"},
		},
		{
			name:       "markdown target with star is compared exactly",
			exactPaths: []string{"logic/*.go"},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectLinkedFiles(files, tt.exactPaths, tt.patterns)
			if !slices.Equal(got, tt.want) {
				t.Errorf("selectLinkedFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompactPromptLinks(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		want   string
	}{
		{
			name:   "body is dropped and footer lists unique links",
			prompt: "body [a](a.go)\n-----\n[[logic/links.go]]\n",
			want:   "-----\n\n[a](a.go)\n[[logic/links.go]]\n",
		},
		{
			name:   "same target in body and footer is listed once",
			prompt: "body [a](a.go)\n-----\n[[a.go]]\n",
			want:   "-----\n\n[a](a.go)\n",
		},
		{
			name:   "links without separator get one in the footer",
			prompt: "body [a](a.go)\n",
			want:   "-----\n\n[a](a.go)\n",
		},
		{
			name:   "prompt without links becomes empty",
			prompt: "plain text\n",
			want:   "",
		},
		{
			name:   "footer keeps only links in order of first appearance",
			prompt: "intro\n-----\nSome note\n[[c.go]] text\n[b](b.go)\n[[c.go]]\n",
			want:   "-----\n\n[[c.go]]\n[b](b.go)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompactPromptLinks(tt.prompt)
			if got != tt.want {
				t.Errorf("CompactPromptLinks(%q) = %q, want %q", tt.prompt, got, tt.want)
			}
		})
	}
}

func TestCompactPromptLinksLeavesOnlyFooter(t *testing.T) {
	compacted := CompactPromptLinks("body [a](a.go)\n-----\n[[logic/links.go]]\n")
	sections := SplitPromptFooter(compacted)

	if sections.Body != "" {
		t.Errorf("SplitPromptFooter(%q).Body = %q, want empty", compacted, sections.Body)
	}
	if sections.Separator != "-----" {
		t.Errorf("SplitPromptFooter(%q).Separator = %q, want %q", compacted, sections.Separator, "-----")
	}
	wantFooter := "[a](a.go)\n[[logic/links.go]]\n"
	if sections.Footer != wantFooter {
		t.Errorf("SplitPromptFooter(%q).Footer = %q, want %q", compacted, sections.Footer, wantFooter)
	}
}
