package logic

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Paths come from LLM output, so anything that would resolve outside the
// project directory must be rejected before it can be written or deleted.
type ProjectPath string

func ParseProjectPath(raw string) (ProjectPath, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("empty file path")
	}

	cleaned := filepath.Clean(strings.TrimPrefix(trimmed, "./"))
	if strings.HasPrefix(cleaned, "/") || filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute path not allowed: %s", raw)
	}

	firstElement := strings.Split(cleaned, string(filepath.Separator))[0]
	if firstElement == ".." {
		return "", fmt.Errorf("path escapes project directory: %s", raw)
	}

	return ProjectPath(cleaned), nil
}