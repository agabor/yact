package api

import (
	"os"
	"strings"
	"yact/config"
	"yact/logic"
)

type Client interface {
	Init(cfg *config.Config)
	GetModelName() string
	Call(transaction logic.Transaction, think bool, systemPrompt string, index []logic.FileEntry) (string, error)
}

func Serialize(path string, index []logic.FileEntry) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	description := getDescriptionFromIndex(path, index)
	delimiter := "//"
	if strings.HasSuffix(path, ".sh") {
		delimiter = "#"
	}

	var headerLine string
	if description != "" {
		headerLine = delimiter + path + "\n" + delimiter + " " + description
	} else {
		headerLine = delimiter + path
	}

	return strings.Join([]string{logic.BlockDelimiter, headerLine, string(content), logic.BlockDelimiter}, "\n"), nil
}

func getDescriptionFromIndex(path string, index []logic.FileEntry) string {
	for _, entry := range index {
		if entry.Path == path {
			return entry.Description
		}
	}
	return ""
}
