package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"yact/logic"
)

func HandleTagCommand(args []string) error {
	if len(args) < 2 {
		fmt.Println("Usage: y tag <tagname> <file> [<file2> ...]")
		return fmt.Errorf("missing arguments")
	}

	tagName := args[0]
	filePatterns := args[1:]

	tags, err := logic.LoadTags()
	if err != nil {
		return err
	}

	for _, pattern := range filePatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("error matching pattern %s: %w", pattern, err)
		}

		if len(matches) == 0 {
			fmt.Printf("No files found matching pattern: %s\n", pattern)
			continue
		}

		for _, filePath := range matches {
			info, err := os.Stat(filePath)
			if err != nil {
				fmt.Printf("Error accessing %s: %v\n", filePath, err)
				continue
			}

			if info.IsDir() {
				fmt.Printf("Skipping directory: %s\n", filePath)
				continue
			}

			if addTagToFile(tags, filePath, tagName) {
				fmt.Printf("Tagged %s with '%s'\n", filePath, tagName)
			} else {
				fmt.Printf("%s already tagged with '%s'\n", filePath, tagName)
			}
		}
	}

	return logic.SaveTags(tags)
}

func addTagToFile(tags map[string][]string, filePath, tagName string) bool {
	existingTags := tags[filePath]
	for _, existingTag := range existingTags {
		if existingTag == tagName {
			return false
		}
	}
	tags[filePath] = append(existingTags, tagName)
	return true
}