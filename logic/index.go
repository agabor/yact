// File indexing and management functions
package logic

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Path        string
	Description string
}

func LoadIndex(filename string) ([]FileEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileEntry{}, nil
		}
		return nil, err
	}

	var entries []FileEntry
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) >= 1 {
			entry := FileEntry{
				Path: unquoteString(parts[0]),
			}
			if len(parts) == 2 {
				entry.Description = unquoteString(parts[1])
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func GetAllFiles(excludePatterns []string) ([]FileEntry, error) {
	var entries []FileEntry
	var excludeDirs = map[string]bool{
		".git":        true,
		".yact":       true,
		"node_modules": true,
		".venv":       true,
		"venv":        true,
		"__pycache__": true,
		"target":      true,
		"dist":        true,
		"build":       true,
	}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		baseName := filepath.Base(path)

		if strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".." {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if excludeDirs[baseName] {
				return filepath.SkipDir
			}
			return nil
		}

		relPath := path
		if strings.HasPrefix(relPath, "./") {
			relPath = relPath[2:]
		}

		shouldExclude := false
		for _, pattern := range excludePatterns {
			matched, err := filepath.Match(pattern, relPath)
			if err == nil && matched {
				shouldExclude = true
				break
			}
		}

		if shouldExclude {
			return nil
		}

		entries = append(entries, FileEntry{
			Path: relPath,
		})

		return nil
	})

	return entries, err
}

func MergeIndex(diskFiles, indexedFiles []FileEntry) []FileEntry {
	descriptionMap := make(map[string]string)
	for _, entry := range indexedFiles {
		if entry.Description != "" {
			descriptionMap[entry.Path] = entry.Description
		}
	}

	var merged []FileEntry
	deletedCount := 0

	for _, diskFile := range diskFiles {
		if description, exists := descriptionMap[diskFile.Path]; exists {
			diskFile.Description = description
		}
		merged = append(merged, diskFile)
	}

	indexMap := make(map[string]bool)
	for _, diskFile := range diskFiles {
		indexMap[diskFile.Path] = true
	}
	for _, indexedFile := range indexedFiles {
		if !indexMap[indexedFile.Path] {
			deletedCount++
		}
	}

	if deletedCount > 0 {
		fmt.Printf("Removed %d deleted file(s) from index\n", deletedCount)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Path < merged[j].Path
	})

	return merged
}

func SaveIndex(filename string, entries []FileEntry) error {
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(quoteString(entry.Path))
		if entry.Description != "" {
			builder.WriteString("  ")
			builder.WriteString(quoteString(entry.Description))
		}
		builder.WriteString("\n")
	}

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}

func quoteString(s string) string {
	return "\"" + s + "\""
}

func unquoteString(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}