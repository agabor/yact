package config

import (
	"reflect"
	"testing"
)

func pointerTo[T any](value T) *T {
	return &value
}

func TestResolveFlags(t *testing.T) {
	tests := []struct {
		name         string
		user         Flaglist
		defaults     PromptDefaults
		wantErr      bool
		wantModel    *ModelName
		wantIndex    *IndexSetting
		wantOutput   OutputName
		wantStrict   bool
		wantDownload bool
	}{
		{
			name:       "defaults fill unset model index and output",
			defaults:   PromptDefaults{Model: pointerTo(ModelHaiku), Index: pointerTo(IndexPublic), Output: pointerTo(OutputNamePlan)},
			wantModel:  pointerTo(ModelHaiku),
			wantIndex:  pointerTo(IndexPublic),
			wantOutput: OutputNamePlan,
		},
		{
			name:       "user model index and output win over defaults",
			user:       Flaglist{Model: pointerTo(ModelOpus), Index: pointerTo(IndexPrivate), Output: pointerTo(OutputNamePrint)},
			defaults:   PromptDefaults{Model: pointerTo(ModelHaiku), Index: pointerTo(IndexPublic), Output: pointerTo(OutputNamePlan)},
			wantModel:  pointerTo(ModelOpus),
			wantIndex:  pointerTo(IndexPrivate),
			wantOutput: OutputNamePrint,
		},
		{
			name:       "explicit index none beats a public default",
			user:       Flaglist{Index: pointerTo(IndexNone)},
			defaults:   PromptDefaults{Index: pointerTo(IndexPublic)},
			wantIndex:  pointerTo(IndexNone),
			wantOutput: OutputNameApply,
		},
		{
			name:       "print output drops a strict default without error",
			user:       Flaglist{Output: pointerTo(OutputNamePrint)},
			defaults:   PromptDefaults{Strict: true},
			wantOutput: OutputNamePrint,
		},
		{
			name:    "strict with print output fails",
			user:    Flaglist{Strict: true, Output: pointerTo(OutputNamePrint)},
			wantErr: true,
		},
		{
			name:    "strict with plan output fails",
			user:    Flaglist{Strict: true, Output: pointerTo(OutputNamePlan)},
			wantErr: true,
		},
		{
			name:    "strict with a print default fails",
			user:    Flaglist{Strict: true},
			defaults: PromptDefaults{Output: pointerTo(OutputNamePrint)},
			wantErr: true,
		},
		{
			name:       "strict default applies when output is apply",
			defaults:   PromptDefaults{Strict: true},
			wantOutput: OutputNameApply,
			wantStrict: true,
		},
		{
			name:         "download default is added",
			defaults:     PromptDefaults{Download: true},
			wantOutput:   OutputNameApply,
			wantDownload: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveFlags(tc.user, tc.defaults)
			if tc.wantErr {
				if err == nil {
					t.Fatal("ResolveFlags() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveFlags() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got.Model, tc.wantModel) {
				t.Errorf("Model = %v, want %v", got.Model, tc.wantModel)
			}
			if !reflect.DeepEqual(got.Index, tc.wantIndex) {
				t.Errorf("Index = %v, want %v", got.Index, tc.wantIndex)
			}
			if got.EffectiveOutput() != tc.wantOutput {
				t.Errorf("EffectiveOutput() = %q, want %q", got.EffectiveOutput(), tc.wantOutput)
			}
			if got.Strict != tc.wantStrict {
				t.Errorf("Strict = %v, want %v", got.Strict, tc.wantStrict)
			}
			if got.Download != tc.wantDownload {
				t.Errorf("Download = %v, want %v", got.Download, tc.wantDownload)
			}
		})
	}
}

func TestParseOutputMode(t *testing.T) {
	tests := []struct {
		name  string
		flags Flaglist
		want  OutputMode
	}{
		{name: "no output set applies", flags: Flaglist{}, want: OutputApply},
		{name: "apply output applies", flags: Flaglist{Output: pointerTo(OutputNameApply)}, want: OutputApply},
		{name: "apply with strict validates and applies", flags: Flaglist{Output: pointerTo(OutputNameApply), Strict: true}, want: OutputValidateAndApply},
		{name: "print output prints", flags: Flaglist{Output: pointerTo(OutputNamePrint)}, want: OutputPrintOnly},
		{name: "plan output writes plan", flags: Flaglist{Output: pointerTo(OutputNamePlan)}, want: OutputWritePlan},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseOutputMode(tc.flags); got != tc.want {
				t.Errorf("ParseOutputMode() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOutputModeNumericValues(t *testing.T) {
	if int(OutputApply) != 0 || int(OutputWritePlan) != 1 || int(OutputPrintOnly) != 2 || int(OutputValidateAndApply) != 3 {
		t.Fatalf("OutputMode values = %d %d %d %d, want 0 1 2 3",
			OutputApply, OutputWritePlan, OutputPrintOnly, OutputValidateAndApply)
	}
}

func TestQueryDefaultsPrintUnlessOutputIsSet(t *testing.T) {
	tests := []struct {
		name string
		user Flaglist
		want OutputName
	}{
		{name: "print without an output flag", user: Flaglist{}, want: OutputNamePrint},
		{name: "output flag wins", user: Flaglist{Output: pointerTo(OutputNamePlan)}, want: OutputNamePlan},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolved, err := ResolveFlags(tc.user, QueryDefaults())
			if err != nil {
				t.Fatalf("ResolveFlags() unexpected error: %v", err)
			}
			if resolved.EffectiveOutput() != tc.want {
				t.Errorf("EffectiveOutput() = %q, want %q", resolved.EffectiveOutput(), tc.want)
			}
		})
	}
}

func TestFlagStringUsesNewNames(t *testing.T) {
	tests := []struct {
		name  string
		flags Flaglist
		want  string
	}{
		{
			name: "named flags",
			flags: Flaglist{
				Model:    pointerTo(ModelHaiku),
				Index:    pointerTo(IndexPrivate),
				Output:   pointerTo(OutputNamePlan),
				Download: true,
				Steps:    "1-3",
				Skills:   []SkillName{SkillName("review")},
			},
			want: "--model haiku --index all --output plan --download --steps 1-3 --skill review",
		},
		{
			name:  "apply output is not printed",
			flags: Flaglist{Output: pointerTo(OutputNameApply)},
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.flags.FlagString(); got != tc.want {
				t.Errorf("FlagString() = %q, want %q", got, tc.want)
			}
		})
	}
}
