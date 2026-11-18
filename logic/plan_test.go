package logic

import (
	"testing"

	"yact/projectpath"
)

func TestAppendReadLinkToPrompt(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		path       string
		want       string
	}{
		{
			name:       "first link into an empty prompt is placed above the separator",
			promptText: "",
			path:       "a.go",
			want:       "[a.go](a.go)\n\n-----\n\n",
		},
		{
			name:       "second link is placed after the first with one empty line between",
			promptText: "[a.go](a.go)\n\n-----\n\n",
			path:       "b.go",
			want:       "[a.go](a.go)\n\n[b.go](b.go)\n\n-----\n\n",
		},
		{
			name:       "body text stays below the separator",
			promptText: "# Plan\nbody\n",
			path:       "a.go",
			want:       "[a.go](a.go)\n\n-----\n\n# Plan\nbody\n",
		},
		{
			name:       "link is added after existing links and above the body",
			promptText: "[a.go](a.go)\n\n-----\n\nbody\n",
			path:       "b.go",
			want:       "[a.go](a.go)\n\n[b.go](b.go)\n\n-----\n\nbody\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := appendReadLinkToPrompt(tc.promptText, tc.path); got != tc.want {
				t.Errorf("appendReadLinkToPrompt(%q, %q) = %q, want %q", tc.promptText, tc.path, got, tc.want)
			}
		})
	}
}

func TestAddReadLinksToPrompt(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		paths      []projectpath.Path
		want       string
	}{
		{
			name:       "empty prompt gets the first link above the separator",
			promptText: "",
			paths:      []projectpath.Path{"a.go"},
			want:       "[a.go](a.go)\n\n-----\n\n",
		},
		{
			name:       "path already linked is not added again",
			promptText: "[a.go](a.go)\n\n-----\n\n",
			paths:      []projectpath.Path{"a.go"},
			want:       "[a.go](a.go)\n\n-----\n\n",
		},
		{
			name:       "several paths keep their order",
			promptText: "",
			paths:      []projectpath.Path{"a.go", "b.go"},
			want:       "[a.go](a.go)\n\n[b.go](b.go)\n\n-----\n\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := AddReadLinksToPrompt(tc.promptText, tc.paths); got != tc.want {
				t.Errorf("AddReadLinksToPrompt(%q, %v) = %q, want %q", tc.promptText, tc.paths, got, tc.want)
			}
		})
	}
}

func TestAppendPromptBody(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		text       string
		want       string
	}{
		{
			name:       "empty prompt returns the text with a trailing newline",
			promptText: "",
			text:       "do x",
			want:       "do x\n",
		},
		{
			name:       "text is appended below the body after a blank line",
			promptText: "# Plan\nbody\n",
			text:       "more",
			want:       "# Plan\nbody\n\nmore\n",
		},
		{
			name:       "read links above the separator stay above it",
			promptText: "[a.go](a.go)\n\n-----\n\nbody\n",
			text:       "more",
			want:       "[a.go](a.go)\n\n-----\n\nbody\n\nmore\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := AppendPromptBody(tc.promptText, tc.text); got != tc.want {
				t.Errorf("AppendPromptBody(%q, %q) = %q, want %q", tc.promptText, tc.text, got, tc.want)
			}
		})
	}
}
