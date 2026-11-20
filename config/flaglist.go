package config

import (
	"fmt"
	"strings"
)

type Flaglist struct {
	Model      *ModelName
	Index      *IndexSetting
	Output     *OutputName
	Batch      bool
	NoProgress bool
	DryRun     bool
}

// OutputApply is the zero value, so a batch job stored without a mode is applied.
// The numeric values are stored in .yact/batch_job_id.txt, so they must not change.
type OutputMode int

const (
	OutputApply OutputMode = iota
	OutputWritePlan
	OutputPrintOnly
)

// User values win over the prompt file defaults.
func ResolveFlags(user Flaglist, defaults PromptDefaults) Flaglist {
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
	resolved.Output = &output
	return resolved
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
	if f.NoProgress {
		parts = append(parts, "--no-progress")
	}
	if f.Batch {
		parts = append(parts, "--batch")
	}

	if f.DryRun {
		parts = append(parts, "--dry-run")
	}

	return strings.Join(parts, " ")
}
