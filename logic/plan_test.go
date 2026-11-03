package logic

import "testing"

func TestAppendPromptLinkToFooter(t *testing.T) {
	tests := []struct {
		name       string
		promptText string
		path       string
		want       string
	}{
		{
			name:       "no separator appends the link to the body",
			promptText: "# Plan\nbody\n",
			path:       "a.go",
			want:       "# Plan\nbody\n=====\n[a.go](a.go)\n",
		},
		{
			name:       "separator with footer appends the link to the footer",
			promptText: "body\n=====\nfooter\n",
			path:       "a.go",
			want:       "body\n=====\nfooter\n[a.go](a.go)\n",
		},
		{
			name:       "separator with empty footer appends the link to the footer",
			promptText: "body\n=====\n",
			path:       "a.go",
			want:       "body\n=====\n[a.go](a.go)\n",
		},
		{
			name:       "footer without trailing line break gets one before the link",
			promptText: "body\n=====\nfooter",
			path:       "a.go",
			want:       "body\n=====\nfooter\n[a.go](a.go)\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := appendPromptLinkToFooter(tc.promptText, tc.path); got != tc.want {
				t.Errorf("appendPromptLinkToFooter(%q, %q) = %q, want %q", tc.promptText, tc.path, got, tc.want)
			}
		})
	}
}
