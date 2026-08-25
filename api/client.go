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
	Call(transaction logic.Transaction, think bool, systemPrompt string) (string, error)
}

func Serialize(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	delimiter := "//"
	if strings.HasSuffix(path, ".sh") {
		delimiter = "#"
	}

	headerLine := delimiter + path

	return strings.Join([]string{logic.BlockDelimiter, headerLine, string(content), logic.BlockDelimiter}, "\n"), nil
}
