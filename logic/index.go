// File indexing and management functions
package logic

import (
	"encoding/csv"
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

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var entries []FileEntry
	for _, record := range records {
		if len(record) == 0 || record[0] == "" {
			continue
		}

		entry := FileEntry{
			Path: record[0],
		}
		if len(record) > 1 {
			entry.Description = record[1]
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func GetAllFiles(excludePatterns []string) ([]FileEntry, error) {
	var entries []FileEntry
	var excludeDirs = map[string]bool{
		".git":         true,
		".yact":        true,
		"node_modules": true,
		".venv":        true,
		"venv":         true,
		"__pycache__":  true,
		"target":       true,
		"dist":         true,
		"build":        true,
	}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		baseName := filepath.Base(path)

		if strings.HasPrefix(baseName, ".") && baseName != "." {
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

		relPath := strings.TrimPrefix(path, "./")

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
	var records [][]string
	for _, entry := range entries {
		record := []string{entry.Path}
		if entry.Description != "" {
			record = append(record, entry.Description)
		}
		records = append(records, record)
	}

	var builder strings.Builder
	w := csv.NewWriter(&builder)
	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}
