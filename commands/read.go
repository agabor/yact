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
			if err := readFile(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func readFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Error accessing %s: %v\n", filePath, err)
		return nil
	}

	if info.IsDir() {
		fmt.Printf("Skipping directory: %s\n", filePath)
		return nil
	}

	requestContent, err := logic.LoadRequestContent()
	if err != nil {
		return err
	}

	if hasFileWithPath(requestContent, filePath) {
		fmt.Printf("Skipping: %s\n", filePath)
		return nil
	}

	fmt.Printf("Reading: %s\n", filePath)

	requestContent.Files = append(requestContent.Files, filePath)

	return requestContent.Save()
}

func hasFileWithPath(requestContent logic.RequestContent, path string) bool {
	for _, contextPath := range requestContent.Files {
		if contextPath == path {
			return true
		}
	}
	return false
}
