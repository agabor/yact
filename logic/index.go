package logic

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Path        string
	Checksum    string
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
		parts := strings.SplitN(line, "  ", 3)
		if len(parts) >= 2 {
			entry := FileEntry{
				Path:     parts[1],
				Checksum: parts[0],
			}
			if len(parts) == 3 {
				entry.Description = parts[2]
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
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

		checksum, err := CalculateChecksum(path)
		if err != nil {
			return nil
		}

		entries = append(entries, FileEntry{
			Path:     relPath,
			Checksum: checksum,
		})

		return nil
	})

	return entries, err
}

func MergeIndex(diskFiles, indexedFiles []FileEntry) []FileEntry {
	indexMap := make(map[string]string)
	descriptionMap := make(map[string]string)
	for _, entry := range indexedFiles {
		indexMap[entry.Path] = entry.Checksum
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
		delete(indexMap, diskFile.Path)
	}

	for _ = range indexMap {
		deletedCount++
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
		builder.WriteString(entry.Checksum)
		builder.WriteString("  ")
		builder.WriteString(entry.Path)
		if entry.Description != "" {
			builder.WriteString("  ")
			builder.WriteString(entry.Description)
		}
		builder.WriteString("\n")
	}

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}