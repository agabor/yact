package logic

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type PromptLinks struct {
	Paths    []string
	Warnings []string
}

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(\s*(?:<([^>]*)>|([^)\s]*))[^)]*\)`)

var urlSchemePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.\-]+:`)

func ResolvePromptLinks(prompt string) PromptLinks {
	targets := extractLinkTargets(prompt)
	links := PromptLinks{}
	seenPaths := make(map[string]bool)

	for _, target := range targets {
		if isURL(target) {
			links.Warnings = append(links.Warnings, fmt.Sprintf("link '%s' skipped: URL not read", target))
			continue
		}

		if err := checkLocalFile(target); err != nil {
			links.Warnings = append(links.Warnings, fmt.Sprintf("link '%s' skipped: %v", target, err))
			continue
		}

		if seenPaths[target] {
			continue
		}

		seenPaths[target] = true
		links.Paths = append(links.Paths, target)
	}

	return links
}

func extractLinkTargets(prompt string) []string {
	targets := make([]string, 0)

	for _, match := range markdownLinkPattern.FindAllStringSubmatch(prompt, -1) {
		target := strings.TrimSpace(match[1])
		if target == "" {
			target = strings.TrimSpace(match[2])
		}
		if target == "" {
			continue
		}
		targets = append(targets, target)
	}

	return targets
}

func isURL(target string) bool {
	if strings.HasPrefix(target, "//") || strings.HasPrefix(target, "#") {
		return true
	}
	return urlSchemePattern.MatchString(target)
}

func checkLocalFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("file unreadable: %v", err)
	}

	if info.IsDir() {
		return fmt.Errorf("is a directory")
	}

	if _, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("file unreadable: %v", err)
	}

	return nil
}
