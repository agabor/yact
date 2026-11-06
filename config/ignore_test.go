package config

import "testing"

func TestParseIgnorePattern(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "root folder name", raw: "vendor", want: "vendor"},
		{name: "nested path", raw: "foo/bar", want: "foo/bar"},
		{name: "single star segment", raw: "foo/*/bar", want: "foo/*/bar"},
		{name: "double star segment", raw: "foo/**/bar", want: "foo/**/bar"},
		{name: "trailing slash is removed", raw: "foo/bar/", want: "foo/bar"},
		{name: "empty pattern", raw: "", wantErr: true},
		{name: "leading slash", raw: "/foo", wantErr: true},
		{name: "empty segment", raw: "foo//bar", wantErr: true},
		{name: "parent segment", raw: "foo/../bar", wantErr: true},
		{name: "backslash", raw: `foo\bar`, wantErr: true},
		{name: "double star inside a segment", raw: "fo**o", wantErr: true},
		{name: "malformed glob", raw: "foo/[", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseIgnorePattern(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseIgnorePattern(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseIgnorePattern(%q) unexpected error: %v", tc.raw, err)
			}
			if got.Text() != tc.want {
				t.Errorf("ParseIgnorePattern(%q).Text() = %q, want %q", tc.raw, got.Text(), tc.want)
			}
		})
	}
}
