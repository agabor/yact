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
	Call(requestContent logic.RequestContent, systemPrompt string) (string, error)
	SubmitBatch(requestContent logic.RequestContent, systemPrompt string) (string, error)
	GetBatchStatus(batchID string) (BatchStatus, error)
	GetBatchResult(batchID string) (BatchResult, error)
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

func BuildUserBlocks(requestContent logic.RequestContent) ([]string, int, error) {
	messagePaths := requestContent.MessagePaths()
	blocks := make([]string, 0, len(messagePaths)+len(requestContent.Files)+1)
	fileCount := 0

	for _, messagePath := range messagePaths {
		messageContent, err := os.ReadFile(messagePath)
		if err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, string(messageContent))
	}

	for _, filePath := range requestContent.Files {
		fileContent, err := Serialize(filePath)
		if err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, fileContent)
		fileCount++
	}

	blocks = append(blocks, requestContent.Prompt)

	return blocks, fileCount, nil
}

func FormatCost(inputTokens int64, outputTokens int64, inputPrice float64, outputPrice float64) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * inputPrice
	outputCost := (float64(outputTokens) / 1_000_000) * outputPrice
	return inputCost + outputCost
}
