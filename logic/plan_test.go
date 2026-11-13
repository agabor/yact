package logic

import "testing"

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
