package config

import "slices"

// The result is returned rather than saved so that the caller writes the file only
// when the extension was actually added.
func WithExtension(extensions []string, extension string) ([]string, bool) {
	if slices.Contains(extensions, extension) {
		return extensions, false
	}

	result := make([]string, 0, len(extensions)+1)
	result = append(result, extensions...)
	result = append(result, extension)
	return result, true
}
