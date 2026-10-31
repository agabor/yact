package index

import "testing"

func TestSkipLineComment(t *testing.T) {
	tests := []struct {
		name   string
		source string
		start  int
		want   int
	}{
		{name: "returns the index of the newline", source: "// a\nb", start: 0, want: 4},
		{name: "returns the length when there is no newline", source: "// a", start: 0, want: 4},
		{name: "starts scanning at the given index", source: "x // c\ny", start: 2, want: 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := skipLineComment(tc.source, tc.start); got != tc.want {
				t.Errorf("skipLineComment(%q, %d) = %d, want %d", tc.source, tc.start, got, tc.want)
			}
		})
	}
}

func TestSkipBlockComment(t *testing.T) {
	tests := []struct {
		name   string
		source string
		start  int
		want   int
	}{
		{name: "returns the index after the closing marker", source: "/* a */x", start: 0, want: 7},
		{name: "unterminated comment runs to the end", source: "/* abc", start: 0, want: 6},
		{name: "starts scanning at the given index", source: "ab/* c */", start: 2, want: 9},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := skipBlockComment(tc.source, tc.start); got != tc.want {
				t.Errorf("skipBlockComment(%q, %d) = %d, want %d", tc.source, tc.start, got, tc.want)
			}
		})
	}
}

func TestSkipLiteral(t *testing.T) {
	tests := []struct {
		name   string
		source string
		start  int
		want   int
	}{
		{name: "escaped quote does not end a double-quoted literal", source: "\"a\\\"b\" x", start: 0, want: 6},
		{name: "backslash does not escape a closing backtick", source: "`a\\` x", start: 0, want: 4},
		{name: "single-quoted literal", source: "'x' y", start: 0, want: 3},
		{name: "unterminated literal runs to the end", source: "\"abc", start: 0, want: 4},
		{name: "starts at the given quote", source: "x = 'y'", start: 4, want: 7},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := skipLiteral(tc.source, tc.start); got != tc.want {
				t.Errorf("skipLiteral(%q, %d) = %d, want %d", tc.source, tc.start, got, tc.want)
			}
		})
	}
}

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "runs of spaces, tabs and newlines become single spaces", text: " a \t b\n\nc  ", want: "a b c"},
		{name: "empty text", text: "", want: ""},
		{name: "single word", text: "x", want: "x"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := collapseWhitespace(tc.text); got != tc.want {
				t.Errorf("collapseWhitespace(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}