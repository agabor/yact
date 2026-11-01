package logic

import (
	"regexp"
	"slices"
	"strings"
)

var promptLinkPattern = regexp.MustCompile(`\[[^\]\n]*\]\((?:<([^>\n]+)>|([^)\s]+))\)`)

type PromptLinks struct {
	Paths []string
}

func linkTarget(match []string) string {
	if match[1] != "" {
		return match[1]
	}
	return match[2]
}

func ParsePromptLinkTargets(prompt string) []string {
	var targets []string
	for _, match := range promptLinkPattern.FindAllStringSubmatch(prompt, -1) {
		targets = append(targets, linkTarget(match))
	}
	return targets
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

func selectLinkedFiles(files []string, targets []string) []string {
	var selected []string
	for _, path := range files {
		if slices.Contains(targets, path) {
			selected = append(selected, path)
		}
	}
	return selected
}

func ResolvePromptLinks(prompt string) (PromptLinks, error) {
	// Only Markdown link targets count, because a plain mention of a path must not send that file to the LLM.
	targets := ParsePromptLinkTargets(prompt)
	if len(targets) == 0 {
		return PromptLinks{}, nil
	}

	files, err := EnumerateFiles(".", func(name string) bool {
		return true
	})
	if err != nil {
		return PromptLinks{}, err
	}

	return PromptLinks{Paths: selectLinkedFiles(files, targets)}, nil
}
