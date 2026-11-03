package logic

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
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

// Identifiers and prose use different separators and casing, so both sides are reduced to the same lowercase words.
func splitWords(text string) []string {
	runes := []rune(text)
	var words []string
	var current []rune
	flush := func() {
		if len(current) > 0 {
			words = append(words, strings.ToLower(string(current)))
			current = nil
		}
	}
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		if i > 0 && isWordBoundary(runes, i) {
			flush()
		}
		current = append(current, r)
	}
	flush()
	return words
}

// An upper-case letter starts a new word after a lower-case letter or digit (camelCase),
// and also when it ends an acronym that is followed by a lower-case letter (HTTPServer).
func isWordBoundary(runes []rune, i int) bool {
	prev, cur := runes[i-1], runes[i]
	if !unicode.IsUpper(cur) {
		return false
	}
	if unicode.IsLower(prev) || unicode.IsDigit(prev) {
		return true
	}
	return unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1])
}

// An empty sequence is not a keyword; matching it would select every file.
func containsWordSequence(words []string, sequence []string) bool {
	if len(sequence) == 0 {
		return false
	}
	for start := 0; start+len(sequence) <= len(words); start++ {
		if slices.Equal(words[start:start+len(sequence)], sequence) {
			return true
		}
	}
	return false
}

// Matching ignores case and separators, so a keyword matches only whole words in order:
// "plan" finds PlanStep and plan_gate but not explanation.
func ContainsKeyword(content string, keyword string) bool {
	return containsWordSequence(splitWords(content), splitWords(keyword))
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
