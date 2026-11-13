package logic

import "testing"

func TestSplitPromptFooter(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		want       PromptSections
	}{
		{
			name:       "no separator keeps the whole text as body",
			promptText: "# Plan\nbody\n",
			want:       PromptSections{Body: "# Plan\nbody\n"},
		},
		{
			name:       "five dashes form the separator",
			promptText: "body\n-----\nfooter\n",
			want:       PromptSections{ReadLinks: "body\n", Body: "footer\n"},
		},
		{
			name:       "blank lines around the separator are not part of read links or body",
			promptText: "body\n\n-----\n\nfooter\n",
			want:       PromptSections{ReadLinks: "body\n", Body: "footer\n"},
		},
		{
			name:       "separator with surrounding spaces is not stored",
			promptText: "body\n   ---  \nfoot",
			want:       PromptSections{ReadLinks: "body\n", Body: "foot"},
		},
		{
			name:       "read links and keywords above the separator are kept",
			promptText: "[a](a.go) ((kw))\n-----\nbody\n",
			want:       PromptSections{ReadLinks: "[a](a.go) ((kw))\n", Body: "body\n"},
		},
		{
			name:       "first separator line is used",
			promptText: "a\n---\nb\n---\nc\n",
			want:       PromptSections{ReadLinks: "a\n", Body: "b\n---\nc\n"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SplitPromptFooter(tc.promptText); got != tc.want {
				t.Errorf("SplitPromptFooter(%q) = %+v, want %+v", tc.promptText, got, tc.want)
			}
		})
	}
}

func TestJoinPromptSections(t *testing.T) {
	tests := []struct {
		name     string
		sections PromptSections
		want     string
	}{
		{
			name:     "sections without read links return the body unchanged",
			sections: PromptSections{Body: "a\nb"},
			want:     "a\nb",
		},
		{
			name:     "read links and body have blank lines around the separator",
			sections: PromptSections{ReadLinks: "a\n", Body: "f\n"},
			want:     "a\n\n-----\n\nf\n",
		},
		{
			name:     "read links without line break get one before the blank line",
			sections: PromptSections{ReadLinks: "a", Body: "f"},
			want:     "a\n\n-----\n\nf",
		},
		{
			name:     "empty body ends after the blank line below the separator",
			sections: PromptSections{ReadLinks: "a\n", Body: ""},
			want:     "a\n\n-----\n\n",
		},
		{
			name:     "round trip of canonical text returns the original text",
			sections: SplitPromptFooter("a\n\n-----\n\nf\n"),
			want:     "a\n\n-----\n\nf\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := JoinPromptSections(tc.sections); got != tc.want {
				t.Errorf("JoinPromptSections(%+v) = %q, want %q", tc.sections, got, tc.want)
			}
		})
	}
}
