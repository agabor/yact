package commands

import "testing"

func TestParseIgnoredFolderArgument(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "empty value returns an error", value: "", wantErr: true},
		{name: "one folder returns that folder", value: "vendor", want: "vendor"},
		{name: "several folders return an error", value: "vendor\nnode_modules", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseIgnoredFolderArgument(tc.value)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseIgnoredFolderArgument(%q) error = nil, want error", tc.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseIgnoredFolderArgument(%q) unexpected error: %v", tc.value, err)
			}
			if got != tc.want {
				t.Errorf("parseIgnoredFolderArgument(%q) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestPrintConfigValue(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		isSensitive bool
		want        string
	}{
		{
			name:        "non-empty sensitive value is masked",
			key:         "anthropic_api_key",
			value:       "secret",
			isSensitive: true,
			want:        "  anthropic_api_key: ******\n",
		},
		{
			name:        "empty sensitive value prints without asterisks",
			key:         "anthropic_api_key",
			value:       "",
			isSensitive: true,
			want:        "  anthropic_api_key: \n",
		},
		{
			name:        "non-sensitive value prints unchanged",
			key:         "claude_model",
			value:       "haiku",
			isSensitive: false,
			want:        "  claude_model: haiku\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				printConfigValue(tc.key, tc.value, tc.isSensitive)
			})
			if got != tc.want {
				t.Errorf("printConfigValue() output = %q, want %q", got, tc.want)
			}
		})
	}
}
