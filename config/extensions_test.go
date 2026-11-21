package config

import (
	"slices"
	"testing"
)

func TestWithExtension(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		extension  string
		want       []string
		wantAdded  bool
	}{
		{name: "adds an extension that is not listed", extensions: []string{"go", "py"}, extension: "rs", want: []string{"go", "py", "rs"}, wantAdded: true},
		{name: "keeps the list when the extension is already listed", extensions: []string{"go", "py"}, extension: "go", want: []string{"go", "py"}, wantAdded: false},
		{name: "adds to an empty list", extensions: []string{}, extension: "go", want: []string{"go"}, wantAdded: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := slices.Clone(tc.extensions)

			got, added := WithExtension(tc.extensions, tc.extension)
			if added != tc.wantAdded {
				t.Errorf("WithExtension(%v, %q) added = %v, want %v", original, tc.extension, added, tc.wantAdded)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("WithExtension(%v, %q) = %v, want %v", original, tc.extension, got, tc.want)
			}
			if !slices.Equal(tc.extensions, original) {
				t.Errorf("WithExtension modified its input: got %v, want %v", tc.extensions, original)
			}
		})
	}
}
