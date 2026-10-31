package index

import "testing"

func TestIsKotlinFunctionLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "function", label: "fun run()", want: true},
		{name: "private function", label: "private fun run()", want: true},
		{name: "suspend function", label: "suspend fun load()", want: true},
		{name: "fun interface", label: "fun interface Runner", want: false},
		{name: "class", label: "class Foo", want: false},
		{name: "property", label: "val x = 1", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isKotlinFunctionLabel(tc.label); got != tc.want {
				t.Errorf("isKotlinFunctionLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsKotlinPublicLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "default visibility is public", label: "fun run()", want: true},
		{name: "private function", label: "private fun run()", want: false},
		{name: "internal function", label: "internal fun x()", want: false},
		{name: "override function", label: "override fun x()", want: true},
		{name: "annotated function", label: "@Test fun x()", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isKotlinPublicLabel(tc.label); got != tc.want {
				t.Errorf("isKotlinPublicLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsSwiftFunctionLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "function", label: "func run()", want: true},
		{name: "public function", label: "public func run()", want: true},
		{name: "attributed function", label: "@objc func run()", want: true},
		{name: "initializer", label: "init(x: Int)", want: true},
		{name: "class method", label: "class func make()", want: true},
		{name: "property", label: "var x = 1", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSwiftFunctionLabel(tc.label); got != tc.want {
				t.Errorf("isSwiftFunctionLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}

func TestIsSwiftPublicLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "public function", label: "public func run()", want: true},
		{name: "open function", label: "open func run()", want: true},
		{name: "internal function", label: "internal func run()", want: false},
		{name: "function without modifier", label: "func run()", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSwiftPublicLabel(tc.label); got != tc.want {
				t.Errorf("isSwiftPublicLabel(%q) = %v, want %v", tc.label, got, tc.want)
			}
		})
	}
}