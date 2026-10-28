package logic

import (
	"regexp"
	"slices"
)

var promptLinkPattern = regexp.MustCompile(`\[[^\]\n]*\]\((?:<([^>\n]+)>|([^)\s]+))\)`)

type PromptLinks struct {
	Paths []string
}

func ParsePromptLinkTargets(prompt string) []string {
	var targets []string
	for _, match := range promptLinkPattern.FindAllStringSubmatch(prompt, -1) {
		if match[1] != "" {
			targets = append(targets, match[1])
		} else {
			targets = append(targets, match[2])
		}
	}
	return targets
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

	files, err := EnumerateFiles(".", nil, func(name string) bool {
		return true
	})
	if err != nil {
		return PromptLinks{}, err
	}

	return PromptLinks{Paths: selectLinkedFiles(files, targets)}, nil
}
