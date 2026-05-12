package commands

import (
	"fmt"
	"strings"
	"yact/api"
	"yact/config"
	"yact/logic"
)

func HandleContextCommand(cfg *config.Config) error {
	fmt.Println("Loading project index...")
	indexedFiles, err := logic.LoadIndex("yact-index.txt")
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
	relevantPaths, err := DiscoverRelevantFiles(indexedFiles, taskPrompt, cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Discovered %d relevant file(s)\n", len(relevantPaths))
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

	transaction := logic.Transaction{
		Context: []string{},
		Request: []string{listingText, taskPrompt},
	}

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	discoveryPrompt, err := config.LoadPrompt("discover")
	if err != nil {
		return nil, fmt.Errorf("error loading system prompt: %w", err)
	}

	response, err := client.Call(transaction, false, discoveryPrompt)
	if err != nil {
		return nil, fmt.Errorf("error discovering relevant files: %w", err)
	}

	filePaths := parseFilePaths(response)
	return filePaths, nil
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
