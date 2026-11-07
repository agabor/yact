package config

import (
	"fmt"
	"strings"
)

type ModelName string

const (
	ModelHaiku  ModelName = "haiku"
	ModelSonnet ModelName = "sonnet"
	ModelOpus   ModelName = "opus"
	ModelFable  ModelName = "fable"
)

func ParseModelName(raw string) (ModelName, error) {
	switch strings.ToLower(raw) {
	case "haiku", "1":
		return ModelHaiku, nil
	case "sonnet", "2":
		return ModelSonnet, nil
	case "opus", "3":
		return ModelOpus, nil
	case "fable", "4":
		return ModelFable, nil
	}
	return "", fmt.Errorf("invalid model %q, expected haiku, sonnet, opus, fable, or 1 to 4", raw)
}

type IndexSetting int

const (
	IndexNone IndexSetting = iota
	IndexPublic
	IndexPrivate
)

func ParseIndexSetting(raw string) (IndexSetting, error) {
	switch strings.ToLower(raw) {
	case "none", "0":
		return IndexNone, nil
	case "public", "1":
		return IndexPublic, nil
	case "all", "2":
		return IndexPrivate, nil
	}
	return IndexNone, fmt.Errorf("invalid index %q, expected none, public, all, or 0 to 2", raw)
}

func (s IndexSetting) Name() string {
	switch s {
	case IndexPublic:
		return "public"
	case IndexPrivate:
		return "all"
	}
	return "none"
}

func (s IndexSetting) Visibility() string {
	switch s {
	case IndexPublic:
		return "public"
	case IndexPrivate:
		return "private"
	}
	return ""
}

type OutputName string

const (
	OutputNameApply OutputName = "apply"
	OutputNamePrint OutputName = "print"
	OutputNamePlan  OutputName = "plan"
)

func ParseOutputName(raw string) (OutputName, error) {
	switch strings.ToLower(raw) {
	case "apply":
		return OutputNameApply, nil
	case "print":
		return OutputNamePrint, nil
	case "plan":
		return OutputNamePlan, nil
	}
	return "", fmt.Errorf("invalid output %q, expected apply, print or plan", raw)
}
