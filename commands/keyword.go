package commands

import (
	"os"
	"yact/config"
	"yact/logic"
	"yact/projectpath"
)

func HandleKeywordCommand(args []string) error {
	files, err := findKeywordFiles(args[0])
	if err != nil {
		return err
	}

	for _, filePath := range files {
		if err := readFile(filePath); err != nil {
			return err
		}
	}

	return nil
}

func findKeywordFiles(keyword string) ([]projectpath.Path, error) {
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

	var matches []projectpath.Path
	for _, path := range files {
		content, err := os.ReadFile(string(path))
		if err != nil {
			continue
		}

		if logic.ContainsKeyword(string(content), keyword) {
			matches = append(matches, path)
		}
	}

	return matches, nil
}

func resolvePromptKeywordFiles(prompt string) ([]projectpath.Path, error) {
	var paths []projectpath.Path
	for _, keyword := range logic.ParsePromptKeywords(prompt) {
		found, err := findKeywordFiles(keyword)
		if err != nil {
			return nil, err
		}
		paths = append(paths, found...)
	}
	return logic.UniquePaths(paths), nil
}
