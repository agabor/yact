package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var includePattern = regexp.MustCompile(`\[\[\s*([^\[\]\s]+)\s*\]\]`)

// PromptName is validated so that it cannot escape the prompts directory or reach a download URL.
type PromptName string

func ParsePromptName(raw string) (PromptName, error) {
	if err := validateIncludeName(raw); err != nil {
		return "", err
	}
	return PromptName(raw), nil
}

func ParseIncludeNames(content string) []string {
	var names []string
	seen := make(map[string]bool)
	for _, match := range includePattern.FindAllStringSubmatch(content, -1) {
		name := match[1]
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

func validateIncludeName(name string) error {
	if name == "" {
		return fmt.Errorf("include name is empty")
	}
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid include name %q", name)
	}
	return nil
}

func ExpandIncludes(content string, readInclude func(name string) (string, error)) (string, error) {
	return expandIncludes(content, readInclude, nil)
}

func expandIncludes(content string, readInclude func(name string) (string, error), chain []string) (string, error) {
	var builder strings.Builder
	last := 0
	for _, location := range includePattern.FindAllStringSubmatchIndex(content, -1) {
		name := content[location[2]:location[3]]
		included, err := expandInclude(name, readInclude, chain)
		if err != nil {
			return "", err
		}
		builder.WriteString(content[last:location[0]])
		builder.WriteString(included)
		last = location[1]
	}
	builder.WriteString(content[last:])
	return builder.String(), nil
}

func expandInclude(name string, readInclude func(name string) (string, error), chain []string) (string, error) {
	if slices.Contains(chain, name) {
		return "", fmt.Errorf("include cycle detected: %s", strings.Join(append(chain, name), " -> "))
	}
	text, err := readInclude(name)
	if err != nil {
		return "", fmt.Errorf("reading include %q: %w", name, err)
	}
	return expandIncludes(text, readInclude, append(chain, name))
}
