package logic

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"yact/projectpath"

	"yact/config"
)

func EnumerateFiles(root string, includeFile func(name string) bool) ([]string, error) {
	ignoredFolders, err := config.LoadIgnoredFolders()
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}

		isRoot := path == root

		if entry.IsDir() {
			if isRoot {
				return nil
			}
			if IsFolderIgnored(relativePath, ignoredFolders) {
				return filepath.SkipDir
			}
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		if IsInIgnoredFolder(relativePath, ignoredFolders) {
			return nil
		}

		if projectpath.Path(entry.Name()) == config.GetProjectPromptPath() {
			return nil
		}

		if includeFile != nil && !includeFile(entry.Name()) {
			return nil
		}

		// Paths are slash-separated on every OS so that links and lookups compare equal on Windows.
		files = append(files, filepath.ToSlash(path))
		return nil
	})
	return files, err
}

// The last segment of a directory path is a folder itself and must be matched
// too, unlike a file path where the last segment is the file name.
func IsFolderIgnored(relativeFolderPath string, ignoredFolders []config.IgnorePattern) bool {
	folders := strings.Split(filepath.ToSlash(relativeFolderPath), "/")
	return matchesAnyFolderPrefix(folders, ignoredFolders)
}

// Patterns are anchored at the project root, and a file is ignored when one of
// the folders leading to it matches, so ignoring "vendor" also hides "vendor/a/b.go".
func IsInIgnoredFolder(relativePath string, ignoredFolders []config.IgnorePattern) bool {
	segments := strings.Split(filepath.ToSlash(relativePath), "/")
	folders := segments[:len(segments)-1]
	return matchesAnyFolderPrefix(folders, ignoredFolders)
}

func matchesAnyFolderPrefix(folders []string, ignoredFolders []config.IgnorePattern) bool {
	for _, pattern := range ignoredFolders {
		for depth := 1; depth <= len(folders); depth++ {
			if matchesFolderSegments(pattern.Segments(), folders[:depth]) {
				return true
			}
		}
	}
	return false
}

// A "**" segment matches zero or more folder segments; every other segment is
// matched with path.Match against exactly one folder segment.
func matchesFolderSegments(pattern []string, folder []string) bool {
	if len(pattern) == 0 {
		return len(folder) == 0
	}
	if pattern[0] == "**" {
		for skip := 0; skip <= len(folder); skip++ {
			if matchesFolderSegments(pattern[1:], folder[skip:]) {
				return true
			}
		}
		return false
	}
	if len(folder) == 0 {
		return false
	}
	matched, err := path.Match(pattern[0], folder[0])
	if err != nil || !matched {
		return false
	}
	return matchesFolderSegments(pattern[1:], folder[1:])
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
