package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"yact/logic"
)

func HandleMessageCommand(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: y message <file> [<file2> ...]")
		return fmt.Errorf("missing file argument")
	}

	for _, pattern := range args {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("error matching pattern %s: %w", pattern, err)
		}

		if len(matches) == 0 {
			fmt.Printf("No files found matching pattern: %s\n", pattern)
			continue
		}

		for _, filePath := range matches {
			if err := addMessageFile(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func addMessageFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Error accessing %s: %v\n", filePath, err)
		return nil
	}

	if info.IsDir() {
		fmt.Printf("Skipping directory: %s\n", filePath)
		return nil
	}

	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	if hasMessageWithPath(transaction, filePath) {
		fmt.Printf("Skipping: %s\n", filePath)
		return nil
	}

	fmt.Printf("Reading: %s\n", filePath)

	transaction.Messages = append(transaction.Messages, filePath)

	return transaction.Save()
}

func hasMessageWithPath(transaction logic.Transaction, path string) bool {
	for _, messagePath := range transaction.Messages {
		if messagePath == path {
			return true
		}
	}
	return false
}