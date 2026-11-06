package config

import (
	"fmt"
	"path"
	"strings"
)

// IgnorePattern is a project-relative folder pattern. Fields are unexported so
// that only ParseIgnorePattern can create one, which keeps every value valid.
type IgnorePattern struct {
	text     string
	segments []string
}

func ParseIgnorePattern(raw string) (IgnorePattern, error) {
	text := strings.TrimSuffix(strings.TrimSpace(raw), "/")
	if text == "" {
		return IgnorePattern{}, fmt.Errorf("ignored folder cannot be empty")
	}
	if strings.HasPrefix(text, "/") {
		return IgnorePattern{}, fmt.Errorf("ignored folder %q must be relative to the project root", raw)
	}
	if strings.Contains(text, `\`) {
		return IgnorePattern{}, fmt.Errorf("ignored folder %q must use forward slashes", raw)
	}

	segments := strings.Split(text, "/")
	for _, segment := range segments {
		if err := validateIgnoreSegment(segment); err != nil {
			return IgnorePattern{}, fmt.Errorf("ignored folder %q: %w", raw, err)
		}
	}

	return IgnorePattern{text: text, segments: segments}, nil
}

func validateIgnoreSegment(segment string) error {
	switch {
	case segment == "":
		return fmt.Errorf("empty path segment")
	case segment == "." || segment == "..":
		return fmt.Errorf("segment %q is not allowed", segment)
	case segment != "**" && strings.Contains(segment, "**"):
		return fmt.Errorf("segment %q may use ** only as a whole segment", segment)
	}

	if _, err := path.Match(segment, ""); err != nil {
		return fmt.Errorf("segment %q is malformed: %w", segment, err)
	}
	return nil
}

func (p IgnorePattern) Text() string {
	return p.text
}

func (p IgnorePattern) Segments() []string {
	return append([]string(nil), p.segments...)
}
