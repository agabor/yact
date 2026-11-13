package commands

import (
	"reflect"
	"testing"

	"yact/config"
)

func TestParseDownloadArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    DownloadTarget
		wantErr bool
	}{
		{
			name: "prompt target",
			args: []string{"prompt", "review"},
			want: PromptDownload{Name: config.PromptName("review")},
		},
		{name: "no arguments", args: []string{}, wantErr: true},
		{name: "prompt without a name", args: []string{"prompt"}, wantErr: true},
		{name: "prompt with two names", args: []string{"prompt", "a", "b"}, wantErr: true},
		{name: "skill is no longer a download target", args: []string{"skill", "review"}, wantErr: true},
		{name: "unknown kind", args: []string{"model", "x"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDownloadArgs(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseDownloadArgs(%v) error = nil, want error", tc.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDownloadArgs(%v) unexpected error: %v", tc.args, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseDownloadArgs(%v) = %#v, want %#v", tc.args, got, tc.want)
			}
		})
	}
}
