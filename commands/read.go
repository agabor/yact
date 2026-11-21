package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"yact/config"
	"yact/logic"
	"yact/projectpath"
)

func HandleReadCommand(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: y read <file> [<file2> ...]")
		return fmt.Errorf("missing file argument")
	}

	ignoredFolders, err := config.LoadIgnoredFolders()
	if err != nil {
		return err
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

		for _, match := range matches {
			filePath, err := projectpath.Parse(match)
			if err != nil {
				fmt.Printf("Rejected: %s: %v\n", match, err)
				continue
			}
			if skipsIgnoredFolder(filePath, ignoredFolders) {
				fmt.Printf("Skipping ignored: %s\n", filePath)
				continue
			}
			if err := readFile(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func skipsIgnoredFolder(filePath projectpath.Path, ignoredFolders []config.IgnorePattern) bool {
	return logic.IsInIgnoredFolder(string(filePath), ignoredFolders)
}

// recordReadExtension keeps the extensions list in step with the files the user
// reads, so that code blocks in later responses for those files are accepted.
func recordReadExtension(filePath projectpath.Path) error {
	extension := logic.FileExtension(string(filePath))
	if extension == "" {
		return nil
	}

	extensions, err := config.LoadExtensions()
	if err != nil {
		return err
	}

	updated, added := config.WithExtension(extensions, extension)
	if !added {
		return nil
	}

	if err := config.SaveExtensions(updated); err != nil {
		return err
	}

	fmt.Printf("Added extension '%s'\n", extension)
	return nil
}

func readFile(filePath projectpath.Path) error {
	info, err := os.Stat(string(filePath))
	if err != nil {
		fmt.Printf("Error accessing %s: %v\n", filePath, err)
		return nil
	}

	if info.IsDir() {
		fmt.Printf("Skipping directory: %s\n", filePath)
		return nil
	}

	if err := recordReadExtension(filePath); err != nil {
		return err
	}

	added, err := logic.AddReadLink(string(filePath))
	if err != nil {
		return err
	}

	if added {
		fmt.Printf("Reading: %s\n", filePath)
	} else {
		fmt.Printf("Skipping: %s\n", filePath)
	}

	return nil
}
