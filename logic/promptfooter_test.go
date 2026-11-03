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
			name:       "five equal signs form the separator",
			promptText: "body\n=====\nfooter\n",
			want:       PromptSections{Body: "body\n", Separator: "=====", Footer: "footer\n"},
		},
		{
			name:       "separator with surrounding spaces is kept verbatim",
			promptText: "body\n   ===  \nfoot",
			want:       PromptSections{Body: "body\n", Separator: "   ===  ", Footer: "foot"},
		},
		{
			name:       "footer with links and keywords is kept",
			promptText: "body\n=====\n[a](a.go) ((kw))\n",
			want:       PromptSections{Body: "body\n", Separator: "=====", Footer: "[a](a.go) ((kw))\n"},
		},
		{
			name:       "first separator line is used",
			promptText: "a\n===\nb\n===\nc\n",
			want:       PromptSections{Body: "a\n", Separator: "===", Footer: "b\n===\nc\n"},
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
			name:     "sections without separator are concatenated",
			sections: PromptSections{Body: "a\n", Footer: "b"},
			want:     "a\nb",
		},
		{
			name:     "sections with separator and footer",
			sections: PromptSections{Body: "a\n", Separator: "===", Footer: "f\n"},
			want:     "a\n===\nf\n",
		},
		{
			name:     "missing line break before separator is added",
			sections: PromptSections{Body: "a", Separator: "===", Footer: "f"},
			want:     "a\n===\nf",
		},
		{
			name:     "round trip returns the original text",
			sections: SplitPromptFooter("a\n===\nf\n"),
			want:     "a\n===\nf\n",
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
