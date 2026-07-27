// Discovers relevant project files for a task using keyword matching and AI-powered file selection
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"yact/config"
	"yact/logic"
)

func HandleContextCommand(cfg *config.Config) error {
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

	fmt.Print("Matching keywords:\n")
	for _, keyword := range keywords {
		fmt.Printf("  - %s\n", keyword)
	}

	keywordMatches := MatchKeywordFiles(keywords)
	if len(keywordMatches) > 0 {
		fmt.Printf("Found %d file(s) by keyword matching\n", len(keywordMatches))
		for _, path := range keywordMatches {
			fmt.Printf("  - %s\n", path)
		}
	}

	transaction.Context = keywordMatches

	if err := transaction.Save(); err != nil {
		return err
	}

	fmt.Println("Context updated successfully")
	return nil
}

func MatchKeywordFiles(keywords []string) []string {
	if len(keywords) == 0 {
		return []string{}
	}

	extensions, err := config.LoadExtensions()
	if err != nil {
		extensions = config.DefaultExtensions()
	}

	var matchedPaths []string
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !hasMatchingExtension(path, extensions) {
			return nil
		}

		fileContent, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentText := string(fileContent)
		for _, keyword := range keywords {
			if strings.Contains(contentText, keyword) {
				matchedPaths = append(matchedPaths, path)
				break
			}
		}

		return nil
	})
	if err != nil {
		return matchedPaths
	}

	return matchedPaths
}

func hasMatchingExtension(path string, extensions []string) bool {
	for _, ext := range extensions {
		suffix := "." + strings.TrimPrefix(ext, ".")
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func extractKeywords(prompt string) []string {
	re := regexp.MustCompile(`\(\(([^)]+)\)\)`)
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
	re := regexp.MustCompile(`\(\([^)]+\)\)`)
	return re.ReplaceAllString(prompt, "")
}
