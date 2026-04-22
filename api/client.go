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

func Serialize(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
	}
	return strings.Join([]string{logic.BlockDelimiter, "//" + path, string(content), logic.BlockDelimiter}, "\n")
}
