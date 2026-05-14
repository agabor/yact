package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"yact/api"
	"yact/config"
	"yact/logic"
)

func HandleContextCommand(cfg *config.Config) error {
	fmt.Println("Loading project index...")
	indexedFiles, err := logic.LoadIndex("yact-index.csv")
	if err != nil {
		return err
	}
	fmt.Printf("Loaded %d files from index\n", len(indexedFiles))

	fmt.Println("Loading current task...")
	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	if len(transaction.Request) == 0 || strings.TrimSpace(transaction.Request[0]) == "" {
		return fmt.Errorf("no task prompt found. Use 'y read' to add context or edit .yact file first")
	}

	taskPrompt := transaction.Request[0]

	fmt.Println("Discovering relevant files for task...")

	keywords := extractKeywords(taskPrompt)
	cleanPrompt := removeKeywordMarkers(taskPrompt)

	keywordMatches := MatchKeywordFiles(indexedFiles, keywords)
	if len(keywordMatches) > 0 {
		fmt.Printf("Found %d file(s) by keyword matching\n", len(keywordMatches))
		for _, path := range keywordMatches {
			fmt.Printf("  - %s (keyword match)\n", path)
		}
	}

	aiMatches, err := DiscoverRelevantFiles(indexedFiles, cleanPrompt, cfg)
	if err != nil {
		return err
	}

	relevantPaths := mergeAndDeduplicatePaths(append(keywordMatches, aiMatches...))

	fmt.Printf("Discovered %d relevant file(s) total\n", len(relevantPaths))
	for _, path := range relevantPaths {
		fmt.Printf("  - %s\n", path)
	}

	transaction.Context = relevantPaths

	if err := transaction.Save(); err != nil {
		return err
	}

	fmt.Println("Context updated successfully")
	return nil
}

func DiscoverRelevantFiles(fileListings []logic.FileEntry, taskPrompt string, cfg *config.Config) ([]string, error) {
	if len(fileListings) == 0 {
		return []string{}, nil
	}

	listingText := BuildFileListingText(fileListings)

	context := []string{}
	if _, err := os.Stat("devdocs.md"); err == nil {
		context = append(context, "devdocs.md")
	}

	transaction := logic.Transaction{
		Context: context,
		Request: []string{listingText, taskPrompt},
	}

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	discoveryPrompt, err := config.LoadPrompt("context")
	if err != nil {
		return nil, fmt.Errorf("error loading system prompt: %w", err)
	}

	indexedFiles, err := logic.LoadIndex("yact-index.csv")
	if err != nil {
		return nil, fmt.Errorf("error loading index: %w", err)
	}

	response, err := client.Call(transaction, false, discoveryPrompt, indexedFiles)
	if err != nil {
		return nil, fmt.Errorf("error discovering relevant files: %w", err)
	}

	filePaths := parseFilePaths(response)
	return filePaths, nil
}

func MatchKeywordFiles(indexedFiles []logic.FileEntry, keywords []string) []string {
	if len(keywords) == 0 {
		return []string{}
	}

	keywordMap := make(map[string]bool)
	for _, keyword := range keywords {
		keywordMap[strings.ToLower(keyword)] = true
	}

	var matchedPaths []string
	for _, entry := range indexedFiles {
		description := strings.ToLower(entry.Description)
		for keyword := range keywordMap {
			if strings.Contains(description, keyword) {
				matchedPaths = append(matchedPaths, entry.Path)
				break
			}
		}
	}

	return matchedPaths
}

func mergeAndDeduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	var merged []string

	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			merged = append(merged, path)
		}
	}

	return merged
}

func BuildFileListingText(entries []logic.FileEntry) string {
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(entry.Path)
		builder.WriteString("\n")
		builder.WriteString(entry.Description)
		builder.WriteString("\n")
		builder.WriteString("\n")
	}
	return builder.String()
}

func parseFilePaths(response string) []string {
	var filePaths []string
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filePaths = append(filePaths, line)
		}
	}

	return filePaths
}

func extractKeywords(prompt string) []string {
	re := regexp.MustCompile(`\[=([^=]+)=\]`)
	matches := re.FindAllStringSubmatch(prompt, -1)

	var keywords []string
	for _, match := range matches {
		if len(match) > 1 {
			keywords = append(keywords, match[1])
		}
	}
	return keywords
}

func removeKeywordMarkers(prompt string) string {
	re := regexp.MustCompile(`\[=[^=]+=\]`)
	return re.ReplaceAllString(prompt, "")
}
