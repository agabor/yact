package commands

import (
	"fmt"
	"regexp"
	"strings"
	"yact/api"
	"yact/config"
	"yact/logic"
)

func HandleIndexCommand(cfg *config.Config) error {
	indexFile := "index.csv"

	fmt.Println("Loading existing index...")
	indexedFiles, err := logic.LoadIndex()
	if err != nil {
		return err
	}
	fmt.Printf("Loaded %d entries from index\n", len(indexedFiles))

	fmt.Println("Scanning disk for files...")
	diskFiles, err := logic.GetAllFiles()
	if err != nil {
		return err
	}
	fmt.Printf("Found %d file(s) on disk\n", len(diskFiles))

	mergedEntries := logic.MergeIndex(diskFiles, indexedFiles)

	fmt.Println("Fetching file descriptions...")
	mergedEntries, err = GetFileDescriptions(mergedEntries, cfg)
	if err != nil {
		fmt.Printf("Warning: Could not fetch descriptions: %v\n", err)
	}

	fmt.Println("Saving index...")
	if err := logic.SaveIndex(indexFile, mergedEntries); err != nil {
		return err
	}

	describedCount := 0
	for _, entry := range mergedEntries {
		if entry.Description != "" {
			describedCount++
		}
	}

	fmt.Printf("Index updated: %s (%d total entries, %d with descriptions)\n", config.GetProjectIndexPath(), len(mergedEntries), describedCount)
	return nil
}

func GetFileDescriptions(entries []logic.FileEntry, cfg *config.Config) ([]logic.FileEntry, error) {
	if len(entries) == 0 {
		return entries, nil
	}

	filePaths := make([]string, 0)
	for _, entry := range entries {
		if entry.Description == "" {
			filePaths = append(filePaths, entry.Path)
		}
	}

	if len(filePaths) == 0 {
		return entries, nil
	}

	fmt.Printf("Fetching descriptions for %d file(s)...\n", len(filePaths))

	transaction := logic.Transaction{
		Context: filePaths,
		Request: []string{"Provide descriptions for these files."},
	}

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	describePrompt, err := config.LoadPrompt("index")
	if err != nil {
		return entries, fmt.Errorf("error loading system prompt: %w", err)
	}

	response, err := client.Call(transaction, false, describePrompt, entries)
	if err != nil {
		return entries, fmt.Errorf("error fetching descriptions: %w", err)
	}

	descriptionMap := parseDescriptions(response)

	updatedEntries := make([]logic.FileEntry, 0)
	for _, entry := range entries {
		if desc, exists := descriptionMap[entry.Path]; exists && desc != "" {
			entry.Description = desc
		}
		updatedEntries = append(updatedEntries, entry)
	}

	return updatedEntries, nil
}

func parseDescriptions(response string) map[string]string {
	descriptionMap := make(map[string]string)

	lines := strings.Split(response, "\n")
	i := 0
	for i < len(lines) {
		currentLine := strings.TrimSpace(lines[i])

		if currentLine == "" {
			i++
			continue
		}

		filePath := unquoteString(currentLine)
		if !isValidFilePath(filePath) {
			i++
			continue
		}

		i++
		if i >= len(lines) {
			break
		}

		descriptionLine := strings.TrimSpace(lines[i])
		if descriptionLine != "" && !isValidFilePath(descriptionLine) {
			descriptionMap[filePath] = unquoteString(descriptionLine)
		}

		i++
	}

	return descriptionMap
}

func isValidFilePath(str string) bool {
	if str == "" || strings.HasPrefix(str, " ") || strings.HasSuffix(str, " ") {
		return false
	}

	if regexp.MustCompile(`^[a-zA-Z0-9._\-/]+$`).MatchString(str) {
		return true
	}

	return false
}

func unquoteString(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}
