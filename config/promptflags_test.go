package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsePromptFlags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    PromptDefaults
	}{
		{
			name:    "letter syntax with model index strict",
			content: "flags: c1vi1\nbody",
			want:    PromptDefaults{Model: pointerTo(ModelHaiku), Index: pointerTo(IndexPublic), Strict: true},
		},
		{
			name:    "letter c without digit falls back to sonnet",
			content: "flags: c",
			want:    PromptDefaults{Model: pointerTo(ModelSonnet)},
		},
		{
			name:    "letter n is print output",
			content: "flags: n",
			want:    PromptDefaults{Output: pointerTo(OutputNamePrint)},
		},
		{
			name:    "letter p is plan output",
			content: "flags: p",
			want:    PromptDefaults{Output: pointerTo(OutputNamePlan)},
		},
		{
			name:    "letter d is download",
			content: "flags: d",
			want:    PromptDefaults{Download: true},
		},
		{
			name:    "long syntax with model strict and index",
			content: "flags: --model haiku --strict --index public",
			want:    PromptDefaults{Model: pointerTo(ModelHaiku), Index: pointerTo(IndexPublic), Strict: true},
		},
		{
			name:    "long syntax with output plan",
			content: "flags: --output plan",
			want:    PromptDefaults{Output: pointerTo(OutputNamePlan)},
		},
		{
			name:    "long syntax with numeric model and download",
			content: "flags: --model 4 --download",
			want:    PromptDefaults{Model: pointerTo(ModelFable), Download: true},
		},
		{
			name:    "content without flags line gives empty defaults",
			content: "just a prompt",
			want:    PromptDefaults{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePromptFlags(PromptName("review"), tc.content)
			if err != nil {
				t.Fatalf("ParsePromptFlags() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParsePromptFlags() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParsePromptFlagsErrorsNameThePrompt(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "unknown flag", content: "flags: --claude 1"},
		{name: "invalid output value", content: "flags: --output maybe"},
		{name: "positional argument", content: "flags: --strict extra"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePromptFlags(PromptName("review"), tc.content)
			if err == nil {
				t.Fatalf("ParsePromptFlags(%q) error = nil, want error", tc.content)
			}
			if !strings.Contains(err.Error(), "review") {
				t.Errorf("error %q does not contain the prompt name", err.Error())
			}
		})
	}
}
