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

type promptLink struct {
	Text   string
	Target string
}

// promptLinkAt records where a link starts, so links of both formats can be put in prompt order.
type promptLinkAt struct {
	start int
	link  promptLink
}

// A glob pattern is only created by ParseGlobPattern, so every GlobPattern is valid for filepath.Match.
type GlobPattern string

func linkTarget(match []string) string {
	if match[1] != "" {
		return match[1]
	}
	return match[2]
}

func submatchStrings(text string, loc []int) []string {
	groups := make([]string, len(loc)/2)
	for i := range groups {
		if loc[2*i] >= 0 {
			groups[i] = text[loc[2*i]:loc[2*i+1]]
		}
	}
	return groups
}

func doubleBracketTargets(prompt string) []string {
	var targets []string
	for _, match := range promptDoubleBracketPattern.FindAllStringSubmatch(prompt, -1) {
		targets = append(targets, match[1])
	}
	return targets
}

func findPromptLinks(prompt string) []promptLink {
	var found []promptLinkAt
	for _, loc := range promptLinkPattern.FindAllStringSubmatchIndex(prompt, -1) {
		match := submatchStrings(prompt, loc)
		found = append(found, promptLinkAt{start: loc[0], link: promptLink{Text: match[0], Target: linkTarget(match)}})
	}
	for _, loc := range promptDoubleBracketPattern.FindAllStringSubmatchIndex(prompt, -1) {
		match := submatchStrings(prompt, loc)
		found = append(found, promptLinkAt{start: loc[0], link: promptLink{Text: match[0], Target: match[1]}})
	}
	// Sorting by position keeps the links in the order they appear in the prompt, which decides which occurrence of a target is kept.
	slices.SortFunc(found, func(a, b promptLinkAt) int { return a.start - b.start })
	links := make([]promptLink, 0, len(found))
	for _, f := range found {
		links = append(links, f.link)
	}
	return links
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
	sections := SplitPromptFooter(prompt)
	links := uniqueLinks(append(findPromptLinks(sections.Body), findPromptLinks(sections.Footer)...))
	if len(links) == 0 {
		return JoinPromptSections(PromptSections{Body: sections.Body})
	}
	if sections.Separator == "" {
		sections.Separator = "-----"
	}
	sections.Footer = linkLines(links)
	return JoinPromptSections(sections)
}

func uniqueLinks(links []promptLink) []promptLink {
	seen := map[string]bool{}
	var unique []promptLink
	for _, link := range links {
		if seen[link.Target] {
			continue
		}
		seen[link.Target] = true
		unique = append(unique, link)
	}
	return unique
}

func linkLines(links []promptLink) string {
	var lines strings.Builder
	for _, link := range links {
		lines.WriteString(link.Text + "\n")
	}
	return lines.String()
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
