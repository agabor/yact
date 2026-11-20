package api

import (
	"strings"
	"yact/config"
	"yact/logic"
)

type Client interface {
	Init(cfg *config.Config)
	GetModelName() string
	Call(requestContent *logic.RequestContent, systemPrompt string) (CallResult, error)
	SubmitBatch(requestContent *logic.RequestContent, systemPrompt string) (string, error)
	GetBatchStatus(batchID string) (BatchStatus, error)
	GetBatchResult(batchID string) (BatchResult, error)
}

func Serialize(file logic.CodeFile) string {
	delimiter := "//"
	if strings.HasSuffix(string(file.Path), ".sh") {
		delimiter = "#"
	}

	headerLine := delimiter + string(file.Path)

	return strings.Join([]string{logic.BlockDelimiter, headerLine, file.Content, logic.BlockDelimiter}, "\n")
}

func BuildUserBlocks(requestContent *logic.RequestContent) ([]string, error) {
	blocks := make([]string, 0, len(requestContent.Files)+2)

	for _, file := range append([]logic.CodeFile{requestContent.Domain, requestContent.Index}, requestContent.Files...) {
		if strings.TrimSpace(file.Content) != "" {
			blocks = append(blocks, Serialize(file))
		}
	}

	if strings.TrimSpace(requestContent.Prompt) != "" {
		blocks = append(blocks, requestContent.Prompt)
	}

	return blocks, nil
}

func FormatCost(inputTokens int64, outputTokens int64, inputPrice float64, outputPrice float64) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * inputPrice
	outputCost := (float64(outputTokens) / 1_000_000) * outputPrice
	return inputCost + outputCost
}
