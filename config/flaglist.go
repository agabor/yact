package config

import (
	"fmt"
	"strings"
)

type Flaglist struct {
	NoWrite       bool
	NoProgress    bool
	ValidateCode  bool
	WritePlan     bool
	Batch         bool
	Download      bool
	Verbose       bool
	Dry           bool
	ClaudeLevel   int
	MaxInputLines int
	Steps         string
	IndexLevel    int
}

// OutputApply is the zero value, so a batch job stored without a mode is applied.
type OutputMode int

const (
	OutputApply OutputMode = iota
	OutputWritePlan
	OutputPrintOnly
	OutputValidateAndApply
)

// Contradictory flags are rejected here so that the order of flags never decides the mode.
func ParseOutputMode(flags Flaglist) (OutputMode, error) {
	if flags.WritePlan && flags.NoWrite {
		return 0, fmt.Errorf("--plan and --no-write cannot be combined")
	}
	if flags.WritePlan && flags.ValidateCode {
		return 0, fmt.Errorf("--plan and --validate-code cannot be combined")
	}
	if flags.NoWrite && flags.ValidateCode {
		return 0, fmt.Errorf("--no-write and --validate-code cannot be combined")
	}
	switch {
	case flags.WritePlan:
		return OutputWritePlan, nil
	case flags.NoWrite:
		return OutputPrintOnly, nil
	case flags.ValidateCode:
		return OutputValidateAndApply, nil
	}
	return OutputApply, nil
}

func (f Flaglist) ModelOverride() string {
	switch f.ClaudeLevel {
	case 1:
		return "haiku"
	case 2:
		return "sonnet"
	case 3:
		return "opus"
	case 4:
		return "fable"
	}
	return ""
}

func (f Flaglist) HasUserSetModel() bool {
	return f.ClaudeLevel > 0
}

func (f Flaglist) HasIndex() bool {
	return f.IndexLevel == 1 || f.IndexLevel == 2
}

func (f Flaglist) FlagString() string {
	var parts []string

	if f.NoWrite {
		parts = append(parts, "--no-write")
	}
	if f.ClaudeLevel > 0 {
		parts = append(parts, fmt.Sprintf("--claude %d", f.ClaudeLevel))
	}
	if f.IndexLevel > 0 {
		parts = append(parts, fmt.Sprintf("--index %d", f.IndexLevel))
	}
	if f.Download {
		parts = append(parts, "--download")
	}
	if f.NoProgress {
		parts = append(parts, "--no-progress")
	}
	if f.ValidateCode {
		parts = append(parts, "--validate-code")
	}
	if f.WritePlan {
		parts = append(parts, "--plan")
	}
	if f.Batch {
		parts = append(parts, "--batch")
	}
	if f.Verbose {
		parts = append(parts, "--verbose")
	}
	if f.Dry {
		parts = append(parts, "--dry")
	}
	if f.MaxInputLines > 0 {
		parts = append(parts, fmt.Sprintf("--max-input-lines %d", f.MaxInputLines))
	}
	if f.Steps != "" {
		parts = append(parts, fmt.Sprintf("--steps %s", f.Steps))
	}

	return strings.Join(parts, " ")
}

func (f *Flaglist) ApplyDefaults(flags []string, skipModelFlags bool) {
	for _, flag := range flags {
		f.applyDefault(flag, skipModelFlags)
	}
}

func (f *Flaglist) applyDefault(flag string, skipModelFlags bool) {
	if strings.HasPrefix(flag, "c") {
		if !skipModelFlags {
			f.ClaudeLevel = claudeLevelFromFlag(flag)
		}
		return
	}

	if strings.HasPrefix(flag, "i") {
		if f.IndexLevel == 0 {
			f.IndexLevel = indexLevelFromFlag(flag)
		}
		return
	}

	switch flag {
	case "n":
		f.NoWrite = true
	case "d":
		f.Download = true
	case "v":
		f.ValidateCode = true
	case "p":
		f.WritePlan = true
	}
}

func claudeLevelFromFlag(flag string) int {
	const fallbackClaudeLevel = 2
	if len(flag) < 2 {
		return fallbackClaudeLevel
	}
	digit := flag[1]
	if digit < '1' || digit > '4' {
		return fallbackClaudeLevel
	}
	return int(digit - '0')
}

func indexLevelFromFlag(flag string) int {
	const fallbackIndexLevel = 1
	if len(flag) < 2 {
		return fallbackIndexLevel
	}
	digit := flag[1]
	if digit != '1' && digit != '2' {
		return fallbackIndexLevel
	}
	return int(digit - '0')
}
