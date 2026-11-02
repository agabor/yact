package commands

import (
	"os"
	"yact/config"
	"yact/logic"
)

func HandleKeywordCommand(args []string) error {
	files, err := findKeywordFiles(args[0])
	if err != nil {
		return err
	}

	for _, path := range files {
		if err := readFile(path); err != nil {
			return err
		}
	}

	return nil
}

func findKeywordFiles(keyword string) ([]string, error) {
	extensions, err := config.LoadExtensions()
	if err != nil {
		return nil, err
	}

	files, err := logic.EnumerateFiles(".", func(name string) bool {
		return logic.FileHasExtension(name, extensions)
	})
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if logic.ContainsKeyword(string(content), keyword) {
			matches = append(matches, path)
		}
	}

	return matches, nil
}

func resolvePromptKeywordFiles(prompt string) ([]string, error) {
	var paths []string
	for _, keyword := range logic.ParsePromptKeywords(prompt) {
		found, err := findKeywordFiles(keyword)
		if err != nil {
			return nil, err
		}
		paths = append(paths, found...)
	}
	return logic.UniquePaths(paths), nil
}
