package logic

import (
	"strings"
)

type PromptLinks struct {
	Paths []string
}

func ResolvePromptLinks(prompt string) (PromptLinks, error) {
	files, err := EnumerateFiles(".", nil, func(name string) bool {
		return true
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
