package logic

import (
	"fmt"
	"path"
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

	// LLM output may use Windows separators; one slash form keeps the same
	// file resolving identically on every OS.
	slashed := strings.ReplaceAll(trimmed, "\\", "/")
	cleaned := path.Clean(strings.TrimPrefix(slashed, "./"))
	if strings.HasPrefix(cleaned, "/") || hasDriveLetter(cleaned) {
		return "", fmt.Errorf("absolute path not allowed: %s", raw)
	}

	firstElement := strings.Split(cleaned, "/")[0]
	if firstElement == ".." {
		return "", fmt.Errorf("path escapes project directory: %s", raw)
	}

	return ProjectPath(cleaned), nil
}

func hasDriveLetter(cleaned string) bool {
	return len(cleaned) >= 2 && cleaned[1] == ':'
}
