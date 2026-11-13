package config

import (
	"fmt"
	"strings"
)

type Flaglist struct {
	Model      *ModelName
	Index      *IndexSetting
	Output     *OutputName
	Strict     bool
	Batch      bool
	NoProgress bool
	Verbose    bool
	DryRun     bool
	Steps      string
}

// OutputApply is the zero value, so a batch job stored without a mode is applied.
// The numeric values are stored in .yact/batch_job_id.txt, so they must not change.
type OutputMode int

const (
	OutputApply OutputMode = iota
	OutputWritePlan
	OutputPrintOnly
	OutputValidateAndApply
)

// User values win over the prompt file defaults. Strict is rejected unless the output is apply,
// so that a conflict is reported instead of silently ignored.
func ResolveFlags(user Flaglist, defaults PromptDefaults) (Flaglist, error) {
	resolved := user
	if resolved.Model == nil {
		resolved.Model = defaults.Model
	}
	if resolved.Index == nil {
		resolved.Index = defaults.Index
	}
	if resolved.Output == nil {
		resolved.Output = defaults.Output
	}

	output := resolved.EffectiveOutput()
	if user.Strict && output != OutputNameApply {
		return Flaglist{}, fmt.Errorf("--strict cannot be combined with --output %s", output)
	}
	resolved.Output = &output
	resolved.Strict = user.Strict || (defaults.Strict && output == OutputNameApply)
	return resolved, nil
}

// Query prints by default, because its response is for reading rather than for applying to files.
func QueryDefaults() PromptDefaults {
	output := OutputNamePrint
	return PromptDefaults{Output: &output}
}

func ParseOutputMode(flags Flaglist) OutputMode {
	switch flags.EffectiveOutput() {
	case OutputNamePrint:
		return OutputPrintOnly
	case OutputNamePlan:
		return OutputWritePlan
	}
	if flags.Strict {
		return OutputValidateAndApply
	}
	return OutputApply
}

func (f Flaglist) EffectiveOutput() OutputName {
	if f.Output == nil {
		return OutputNameApply
	}
	return *f.Output
}

func (f Flaglist) ModelOverride() string {
	if f.Model == nil {
		return ""
	}
	return string(*f.Model)
}

// --output plan sends the plan whole, so it implies selector 0 unless --steps is given explicitly.
func (f Flaglist) StepsSelector() string {
	if f.EffectiveOutput() == OutputNamePlan && f.Steps == "" {
		return "0"
	}
	return f.Steps
}

func (f Flaglist) FlagString() string {
	var parts []string

	if f.Model != nil {
		parts = append(parts, fmt.Sprintf("--model %s", *f.Model))
	}
	if f.Index != nil {
		parts = append(parts, fmt.Sprintf("--index %s", f.Index.Name()))
	}
	if f.Output != nil && *f.Output != OutputNameApply {
		parts = append(parts, fmt.Sprintf("--output %s", *f.Output))
	}
	if f.Strict {
		parts = append(parts, "--strict")
	}
	if f.NoProgress {
		parts = append(parts, "--no-progress")
	}
	if f.Batch {
		parts = append(parts, "--batch")
	}
	if f.Verbose {
		parts = append(parts, "--verbose")
	}
	if f.DryRun {
		parts = append(parts, "--dry-run")
	}
	if f.Steps != "" {
		parts = append(parts, fmt.Sprintf("--steps %s", f.Steps))
	}

	return strings.Join(parts, " ")
}
