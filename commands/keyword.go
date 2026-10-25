package commands

import (
	"os"
	"strings"
	"yact/config"
	"yact/logic"
)

func HandleKeywordCommand(args []string) error {
	keyword := args[0]

	extensions, err := config.LoadExtensions()
	if err != nil {
		return err
	}

	files, err := logic.EnumerateFiles(".", nil, func(name string) bool {
		return logic.FileHasExtension(name, extensions)
	})
	if err != nil {
		return err
	}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if !strings.Contains(string(content), keyword) {
			continue
		}

		if err := readFile(path); err != nil {
			return err
		}
	}

	return nil
}
