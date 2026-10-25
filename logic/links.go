package logic

import (
	"strings"
	"yact/config"
)

type PromptLinks struct {
	Paths []string
}

func ResolvePromptLinks(prompt string) (PromptLinks, error) {
	extensions, err := config.LoadExtensions()
	if err != nil {
		return PromptLinks{}, err
	}

	files, err := EnumerateFiles(".", nil, func(name string) bool {
		return FileHasExtension(name, extensions)
	})
	if err != nil {
		return PromptLinks{}, err
	}

	links := PromptLinks{}
	for _, path := range files {
		if strings.Contains(prompt, path) {
			links.Paths = append(links.Paths, path)
		}
	}

	return links, nil
}
