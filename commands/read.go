package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"yact/logic"
)

func HandleReadCommand(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: y read <file> [<file2> ...]")
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
			info, err := os.Stat(filePath)
			if err != nil {
				fmt.Printf("Error accessing %s: %v\n", filePath, err)
				continue
			}

			if info.IsDir() {
				fmt.Printf("Skipping directory: %s\n", filePath)
				continue
			}

			transaction, err := logic.LoadContext()
			if err != nil {
				return err
			}

			if hasFileWithPath(transaction, filePath) {
				fmt.Printf("Skipping: %s\n", filePath)
				continue
			} else {
				fmt.Printf("Reading: %s\n", filePath)
			}

			transaction.Context = append(transaction.Context, filePath)

			err2 := logic.SaveContext(transaction)
			if err2 != nil {
				return err2
			}
		}
	}

	return nil
}

func hasFileWithPath(transaction logic.Transaction, path string) bool {
	for _, contextPath := range transaction.Context {
		if contextPath == path {
			return true
		}
	}
	return false
}