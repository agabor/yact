// Defines the filesystem path for the project stash file
package config

import "path/filepath"

func GetProjectStashPath() string {
	return filepath.Join(".yact", "stash.txt")
}