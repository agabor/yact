package logic

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func EnumerateFiles(root string, skipDir func(name string) bool, includeFile func(name string) bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		isRoot := path == root
		if !isRoot && strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			if !isRoot && skipDir != nil && skipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		if includeFile != nil && !includeFile(entry.Name()) {
			return nil
		}

		files = append(files, path)
		return nil
	})
	return files, err
}

func FileHasExtension(path string, extensions []string) bool {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	for _, allowed := range extensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
