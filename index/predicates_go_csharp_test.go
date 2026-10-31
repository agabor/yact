package index

import "testing"

func TestIsGoFunctionLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "named function", label: "func Foo()", want: true},
		{name: "function literal", label: "func(x int)", want: true},
		{name: "identifier starting with func", label: "funcs", want: false},
		{name: "type declaration", label: "type T struct", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isGoFunctionLabel(tc.label); got != tc.want {
				t.Errorf("isGoFunctionLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsGoPublicLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "exported function", label: "func Foo()", want: true},
		{name: "unexported function", label: "func foo()", want: false},
		{name: "exported method", label: "func (r *T) Bar()", want: true},
		{name: "unexported method", label: "func (r *T) bar()", want: false},
		{name: "bare func keyword", label: "func", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isGoPublicLabel(tc.label); got != tc.want {
				t.Errorf("isGoPublicLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsCSharpFunctionLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "method with return type", label: "public void Run()", want: true},
		{name: "constructor", label: "public Foo(int x)", want: true},
		{name: "control statement", label: "if (x)", want: false},
		{name: "assignment with call", label: "int x = Compute();", want: false},
		{name: "bare call", label: "Run()", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCSharpFunctionLabel(tc.label); got != tc.want {
				t.Errorf("isCSharpFunctionLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsCSharpPublicLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "public static method", label: "public static void X()", want: true},
		{name: "internal method", label: "internal void X()", want: false},
		{name: "attribute before public method", label: "[Obsolete] public void X()", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The scanner strips attributes before the predicate runs, so the test does the same.
			label := cleanCSharpLabel(tc.label)
			if got := isCSharpPublicLabel(label); got != tc.want {
				t.Errorf("isCSharpPublicLabel(%q) = %v, want %v", label, got, tc.want)
			}
		})
	}
}