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
	Call(transaction logic.Transaction, cfg *config.Config, systemPrompt string) (string, error)
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

func BuildUserBlocks(transaction logic.Transaction) ([]string, int, error) {
	blocks := make([]string, 0, len(transaction.Context)+len(transaction.Request))
	fileCount := 0

	for _, filePath := range transaction.Context {
		fileContent, err := Serialize(filePath)
		if err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, fileContent)
		fileCount++
	}

	for _, prompt := range transaction.Request {
		blocks = append(blocks, prompt)
	}

	return blocks, fileCount, nil
}

func FormatCost(inputTokens int64, outputTokens int64, inputPrice float64, outputPrice float64) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * inputPrice
	outputCost := (float64(outputTokens) / 1_000_000) * outputPrice
	return inputCost + outputCost
}