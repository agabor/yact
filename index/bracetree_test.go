package index

import (
	"reflect"
	"strings"
	"testing"
)

func flattenBraceTree(node *braceNode) []string {
	return appendBraceLabels(nil, node, 0)
}

func appendBraceLabels(lines []string, node *braceNode, depth int) []string {
	for _, child := range node.Children {
		lines = append(lines, strings.Repeat("  ", depth)+child.Label)
		lines = appendBraceLabels(lines, child, depth+1)
	}
	return lines
}

func TestParseBraceTreeNesting(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "nested block is a child of the block around it",
			source: "func Outer() {\n\tfunc Inner() {\n\t}\n}\n",
			want:   []string{"func Outer()", "  func Inner()"},
		},
		{
			name:   "sibling blocks share the same parent",
			source: "func A() {\n}\nfunc B() {\n}\n",
			want:   []string{"func A()", "func B()"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenBraceTree(parseBraceTree(tc.source, false))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("tree = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseBraceTreeLineBreakSplitting(t *testing.T) {
	tests := []struct {
		name               string
		source             string
		splitsAtLineBreak  bool
		want               []string
	}{
		{
			name:              "label on its own line is cut off when splitting",
			source:            "x := 1\nfunc Run() {\n}\n",
			splitsAtLineBreak: true,
			want:              []string{"func Run()"},
		},
		{
			name:              "label stays joined when not splitting",
			source:            "x := 1\nfunc Run() {\n}\n",
			splitsAtLineBreak: false,
			want:              []string{"x := 1 func Run()"},
		},
		{
			name:              "line ending after a comma continues the label",
			source:            "a,\nfunc Run() {\n}\n",
			splitsAtLineBreak: true,
			want:              []string{"a, func Run()"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenBraceTree(parseBraceTree(tc.source, tc.splitsAtLineBreak))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("tree = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseBraceTreeKeepsLabelsInsideParentheses(t *testing.T) {
	source := "func Run(a int\nb int) {\n}\n"
	got := flattenBraceTree(parseBraceTree(source, true))
	want := []string{"func Run(a int b int)"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tree = %q, want %q", got, want)
	}
}

func TestParseBraceTreeIgnoresBracesInLiteralsAndComments(t *testing.T) {
	tests := []struct {
		name              string
		source            string
		splitsAtLineBreak bool
		want              []string
	}{
		{
			name:              "semicolon resets the label and braces in literals create no node",
			source:            "var s = \"{\"; func Run() {}\n",
			splitsAtLineBreak: false,
			want:              []string{"func Run()"},
		},
		{
			name:              "braces in a line comment create no node",
			source:            "// func Hidden() {\nfunc Shown() {\n}\n",
			splitsAtLineBreak: false,
			want:              []string{"func Shown()"},
		},
		{
			name:              "braces in a block comment create no node",
			source:            "/* func Hidden() { */ func Shown() {\n}\n",
			splitsAtLineBreak: false,
			want:              []string{"func Shown()"},
		},
		{
			name:              "quote in a label does not end the label early",
			source:            "x := \"}\"\nfunc Run() {\n}\n",
			splitsAtLineBreak: true,
			want:              []string{"func Run()"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenBraceTree(parseBraceTree(tc.source, tc.splitsAtLineBreak))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("tree = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractLabel(t *testing.T) {
	tests := []struct {
		name   string
		source string
		start  int
		end    int
		want   string
	}{
		{
			name:   "comments are removed",
			source: "// note\nfunc A() ",
			start:  0,
			end:    len("// note\nfunc A() "),
			want:   "func A()",
		},
		{
			name:   "block comment is removed",
			source: "func A(/* x */) ",
			start:  0,
			end:    len("func A(/* x */) "),
			want:   "func A()",
		},
		{
			name:   "literal contents are kept",
			source: "msg := \"hi // not\"",
			start:  0,
			end:    len("msg := \"hi // not\""),
			want:   "msg := \"hi // not\"",
		},
		{
			name:   "whitespace is collapsed",
			source: "func   A(\n\tx int)  ",
			start:  0,
			end:    len("func   A(\n\tx int)  "),
			want:   "func A( x int)",
		},
		{
			name:   "range limits the label",
			source: "skip func A() {",
			start:  5,
			end:    13,
			want:   "func A()",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractLabel(tc.source, tc.start, tc.end); got != tc.want {
				t.Errorf("extractLabel() = %q, want %q", got, tc.want)
			}
		})
	}
}