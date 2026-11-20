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
			name:    "letter syntax with model and index",
			content: "flags: c1i1\nbody",
			want:    PromptDefaults{Model: pointerTo(ModelHaiku), Index: pointerTo(IndexPublic)},
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
			name:    "long syntax with model and index",
			content: "flags: --model haiku --index public",
			want:    PromptDefaults{Model: pointerTo(ModelHaiku), Index: pointerTo(IndexPublic)},
		},
		{
			name:    "long syntax with output plan",
			content: "flags: --output plan",
			want:    PromptDefaults{Output: pointerTo(OutputNamePlan)},
		},
		{
			name:    "long syntax with numeric model",
			content: "flags: --model 4",
			want:    PromptDefaults{Model: pointerTo(ModelFable)},
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
		{name: "positional argument", content: "flags: --output plan extra"},
		{name: "download is no longer a flag", content: "flags: --download"},
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

func TestStripPromptFlagsLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "content without flags line is unchanged",
			content: "just a prompt\nsecond line",
			want:    "just a prompt\nsecond line",
		},
		{
			name:    "flags line is removed before the body",
			content: "flags: c1i1\nbody line",
			want:    "body line",
		},
		{
			name:    "content consisting only of the flags line becomes empty",
			content: "flags: c1",
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := StripPromptFlagsLine(tc.content)
			if got != tc.want {
				t.Errorf("StripPromptFlagsLine(%q) = %q, want %q", tc.content, got, tc.want)
			}
		})
	}
}
