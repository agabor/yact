package logic

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var promptLinkPattern = regexp.MustCompile(`\[[^\]\n]*\]\((?:<([^>\n]+)>|([^)\s]+))\)`)

// A double-bracket link is a shorthand for a Markdown link with the same target, so [[main.go]] sends the same file as [main.go](main.go).
var promptDoubleBracketPattern = regexp.MustCompile(`\[\[([^\[\]\n]+)\]\]`)

type PromptLinks struct {
	Paths []string
}

// A glob pattern is only created by ParseGlobPattern, so every GlobPattern is valid for filepath.Match.
type GlobPattern string

func linkTarget(match []string) string {
	if match[1] != "" {
		return match[1]
	}
	return match[2]
}

func doubleBracketTargets(prompt string) []string {
	var targets []string
	for _, match := range promptDoubleBracketPattern.FindAllStringSubmatch(prompt, -1) {
		targets = append(targets, match[1])
	}
	return targets
}

func ParsePromptLinkTargets(prompt string) []string {
	var targets []string
	for _, match := range promptLinkPattern.FindAllStringSubmatch(prompt, -1) {
		targets = append(targets, linkTarget(match))
	}
	return append(targets, doubleBracketTargets(prompt)...)
}

// Markdown link targets are exact paths, so a glob in them is never expanded.
func ParseMarkdownLinkTargets(prompt string) []string {
	var targets []string
	for _, match := range promptLinkPattern.FindAllStringSubmatch(prompt, -1) {
		targets = append(targets, linkTarget(match))
	}
	return targets
}

func ParseGlobPattern(raw string) (GlobPattern, error) {
	if _, err := filepath.Match(raw, ""); err != nil {
		return "", err
	}
	return GlobPattern(raw), nil
}

func parseGlobPatterns(raws []string) ([]GlobPattern, error) {
	var patterns []GlobPattern
	for _, raw := range raws {
		pattern, err := ParseGlobPattern(raw)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

// The pattern was validated by ParseGlobPattern, so filepath.Match cannot return an error here.
func (g GlobPattern) Matches(path string) bool {
	matched, _ := filepath.Match(string(g), path)
	return matched
}

// Links are compared by target only, so <a> and a are the same link even though they are written differently.
func CompactPromptLinks(prompt string) string {
	seen := map[string]bool{}
	var compacted strings.Builder
	for _, match := range promptLinkPattern.FindAllStringSubmatch(prompt, -1) {
		target := linkTarget(match)
		if seen[target] {
			continue
		}
		seen[target] = true
		compacted.WriteString(match[0] + "\n")
	}
	return compacted.String()
}

func selectLinkedFiles(files []string, exactPaths []string, patterns []GlobPattern) []string {
	var selected []string
	for _, path := range files {
		if slices.Contains(exactPaths, path) || slices.ContainsFunc(patterns, func(pattern GlobPattern) bool {
			return pattern.Matches(path)
		}) {
			selected = append(selected, path)
		}
	}
	return selected
}

func ResolvePromptLinks(prompt string) (PromptLinks, error) {
	// Only link targets count, because a plain mention of a path must not send that file to the LLM.
	exactPaths := ParseMarkdownLinkTargets(prompt)
	patternRaws := doubleBracketTargets(prompt)
	if len(exactPaths) == 0 && len(patternRaws) == 0 {
		return PromptLinks{}, nil
	}

	patterns, err := parseGlobPatterns(patternRaws)
	if err != nil {
		return PromptLinks{}, err
	}

	files, err := EnumerateFiles(".", func(name string) bool {
		return true
	})
	if err != nil {
		return PromptLinks{}, err
	}

	return PromptLinks{Paths: selectLinkedFiles(files, exactPaths, patterns)}, nil
}
