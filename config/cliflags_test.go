package config

import (
	"io"
	"reflect"
	"testing"

	"github.com/spf13/pflag"
)

func newTestFlagSet() (*pflag.FlagSet, *CLIFlags) {
	set := pflag.NewFlagSet("test", pflag.ContinueOnError)
	set.SetOutput(io.Discard)
	return set, RegisterFlags(set)
}

func parseTestArgs(t *testing.T, args []string) Flaglist {
	t.Helper()
	set, cli := newTestFlagSet()
	if err := set.Parse(args); err != nil {
		t.Fatalf("Parse(%v) unexpected error: %v", args, err)
	}
	flags, err := cli.ToFlaglist(set)
	if err != nil {
		t.Fatalf("ToFlaglist() unexpected error: %v", err)
	}
	return flags
}

func TestRemovedFlagsAreUnknown(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "claude long with value", args: []string{"--claude", "1"}},
		{name: "claude short with value", args: []string{"-c1"}},
		{name: "no-write long", args: []string{"--no-write"}},
		{name: "no-write short", args: []string{"-n"}},
		{name: "plan long", args: []string{"--plan"}},
		{name: "plan short", args: []string{"-p"}},
		{name: "validate-code long", args: []string{"--validate-code"}},
		{name: "validate-code short", args: []string{"-v"}},
		{name: "dry long", args: []string{"--dry"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set, _ := newTestFlagSet()
			if err := set.Parse(tc.args); err == nil {
				t.Fatalf("Parse(%v) error = nil, want unknown flag error", tc.args)
			}
		})
	}
}

func TestStepsSelectorAll(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "all becomes selector 0", args: []string{"--steps", "all"}, want: "0"},
		{name: "range is unchanged", args: []string{"--steps", "1-3"}, want: "1-3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseTestArgs(t, tc.args).Steps; got != tc.want {
				t.Errorf("Steps = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestModelFlagConversion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *ModelName
		wantErr bool
	}{
		{name: "name", args: []string{"--model", "opus"}, want: pointerTo(ModelOpus)},
		{name: "number", args: []string{"--model", "3"}, want: pointerTo(ModelOpus)},
		{name: "unset stays nil", args: nil, want: nil},
		{name: "invalid model", args: []string{"--model", "claude"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set, cli := newTestFlagSet()
			if err := set.Parse(tc.args); err != nil {
				t.Fatalf("Parse(%v) unexpected error: %v", tc.args, err)
			}
			flags, err := cli.ToFlaglist(set)
			if tc.wantErr {
				if err == nil {
					t.Fatal("ToFlaglist() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ToFlaglist() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(flags.Model, tc.want) {
				t.Errorf("Model = %v, want %v", flags.Model, tc.want)
			}
		})
	}
}

func TestIndexFlagConversion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want *IndexSetting
	}{
		{name: "all is private", args: []string{"--index", "all"}, want: pointerTo(IndexPrivate)},
		{name: "unset stays nil", args: nil, want: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := parseTestArgs(t, tc.args)
			if !reflect.DeepEqual(flags.Index, tc.want) {
				t.Errorf("Index = %v, want %v", flags.Index, tc.want)
			}
		})
	}
}
