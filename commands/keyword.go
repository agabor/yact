package commands

import (
	"fmt"
	"os"
	"path/filepath"
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

	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if isHiddenDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !fileHasExtension(path, extensions) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if !strings.Contains(string(content), keyword) {
			return nil
		}

		transaction, err := logic.LoadTransaction()
		if err != nil {
			return err
		}

		if hasFileWithPath(transaction, path) {
			fmt.Printf("Skipping: %s\n", path)
			return nil
		}

		fmt.Printf("Reading: %s\n", path)

		transaction.Context = append(transaction.Context, path)

		return transaction.Save()
	})
}

func isHiddenDir(path string) bool {
	if path == "." {
		return false
	}
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".")
}

func fileHasExtension(path string, extensions []string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	for _, allowed := range extensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
