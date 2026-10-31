package index

import "testing"

func TestIsJSFunctionLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "function declaration", label: "function run()", want: true},
		{name: "exported default function", label: "export default function run()", want: true},
		{name: "arrow function binding", label: "const f = () =>", want: true},
		{name: "method", label: "run()", want: true},
		{name: "control statement", label: "if (x)", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isJSFunctionLabel(tc.label); got != tc.want {
				t.Errorf("isJSFunctionLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsJSPublicLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "public method", label: "run()", want: true},
		{name: "private method", label: "#secret()", want: false},
		{name: "exported function", label: "export function f()", want: true},
		{name: "unexported function", label: "function f()", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isJSPublicLabel(tc.label); got != tc.want {
				t.Errorf("isJSPublicLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsPHPFunctionLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "public method", label: "public function run()", want: true},
		{name: "public static method", label: "public static function f()", want: true},
		{name: "function without modifier", label: "function run()", want: true},
		{name: "statement", label: "$x = 1;", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPHPFunctionLabel(tc.label); got != tc.want {
				t.Errorf("isPHPFunctionLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsPHPPublicLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "public method", label: "public function run()", want: true},
		{name: "private method", label: "private function h()", want: false},
		{name: "protected static method", label: "protected static function p()", want: false},
		{name: "function without modifier", label: "function f()", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPHPPublicLabel(tc.label); got != tc.want {
				t.Errorf("isPHPPublicLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}