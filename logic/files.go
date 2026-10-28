package logic

import (
	"io/fs"
	"path/filepath"
	"strings"

	"yact/config"
)

func EnumerateFiles(root string, includeFile func(name string) bool) ([]string, error) {
	ignoredRootFolders, err := config.LoadIgnoredFolders()
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
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
			if isRoot {
				return nil
			}
			relativePath, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if IsInIgnoredRootFolder(relativePath, ignoredRootFolders) {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		if entry.Name() == config.GetProjectPromptPath() {
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

func IsInIgnoredRootFolder(relativePath string, ignoredRootFolders []string) bool {
	firstSegment := strings.SplitN(filepath.ToSlash(relativePath), "/", 2)[0]
	for _, folder := range ignoredRootFolders {
		if folder == firstSegment {
			return true
		}
	}
	return false
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
