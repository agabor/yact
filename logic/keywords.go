package logic

import (
	"regexp"
	"strings"
)

// Phrases are wrapped in double parentheses so that ordinary parentheses in prose do not trigger a file search.
var promptKeywordPattern = regexp.MustCompile(`\(\(([^()\n]+)\)\)`)

func ParsePromptKeywords(prompt string) []string {
	seen := map[string]bool{}
	var keywords []string
	for _, match := range promptKeywordPattern.FindAllStringSubmatch(prompt, -1) {
		phrase := strings.TrimSpace(match[1])
		if phrase == "" || seen[phrase] {
			continue
		}
		seen[phrase] = true
		keywords = append(keywords, phrase)
	}
	return keywords
}

// The LLM must not see the markers, but the keyword text itself is still useful.
func StripKeywordMarkers(prompt string) string {
	return promptKeywordPattern.ReplaceAllString(prompt, "$1")
}

// The match is case-sensitive, the same rule the keyword command has always used.
func ContainsKeyword(content string, keyword string) bool {
	return strings.Contains(content, keyword)
}

func UniquePaths(paths []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, path := range paths {
		if seen[path] {
			continue
		}
		seen[path] = true
		unique = append(unique, path)
	}
	return unique
}
