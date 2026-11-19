package api

import (
	"os"
	"strings"
	"yact/config"
	"yact/logic"
	"yact/projectpath"
)

type Client interface {
	Init(cfg *config.Config)
	GetModelName() string
	Call(requestContent logic.RequestContent, systemPrompt string) (CallResult, error)
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

func BuildUserBlocks(requestContent logic.RequestContent) ([]string, error) {
	blocks := make([]string, 0, len(requestContent.Files)+2)

	for _, filePath := range append([]projectpath.Path{requestContent.Domain, requestContent.Index}, requestContent.Files...) {
		fileContent, err := Serialize(string(filePath))
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, fileContent)
	}

	blocks = append(blocks, requestContent.Prompt)

	return blocks, nil
}

func FormatCost(inputTokens int64, outputTokens int64, inputPrice float64, outputPrice float64) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * inputPrice
	outputCost := (float64(outputTokens) / 1_000_000) * outputPrice
	return inputCost + outputCost
}
