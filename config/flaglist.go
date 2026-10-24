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
	ClaudeLevel   int
	MaxInputLines int
	Steps         string
	PromptFile    string
	IndexLevel    int
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
	if f.MaxInputLines > 0 {
		parts = append(parts, fmt.Sprintf("--max-input-lines %d", f.MaxInputLines))
	}
	if f.Steps != "" {
		parts = append(parts, fmt.Sprintf("--steps %s", f.Steps))
	}
	if f.PromptFile != "" {
		parts = append(parts, fmt.Sprintf("--file %s", f.PromptFile))
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

	switch flag {
	case "n":
		if !f.ValidateCode {
			f.NoWrite = true
		}
	case "d":
		f.Download = true
	case "v":
		if !f.NoWrite {
			f.ValidateCode = true
		}
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