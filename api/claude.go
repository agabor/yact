package api

import (
	"context"
	"fmt"
	"strings"
	"time"
	"yact/config"
	"yact/logic"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const haikuLongPromptThreshold int64 = 100000

const batchCustomID = "yact-request"

const batchCostMultiplier = 0.5

const batchEndedStatus = "ended"

type BatchStatus struct {
	ProcessingStatus string
	Ended            bool
	Processing       int64
	Succeeded        int64
	Errored          int64
	Canceled         int64
	Expired          int64
}

type BatchResult struct {
	Text         string
	InputTokens  int64
	OutputTokens int64
	Cost         float64
	Failure      string
}

type claudeTier struct {
	model           anthropic.Model
	inputPrice      float64
	outputPrice     float64
	longInputPrice  float64
	longOutputPrice float64
}

type ClaudeClient struct {
	apiKey string
	tier   claudeTier
}

func (c *ClaudeClient) Init(cfg *config.Config) {
	c.apiKey = cfg.AnthropicAPIKey
	c.tier = selectTier(cfg.ClaudeModel)
}

func selectTier(modelName string) claudeTier {
	switch strings.ToLower(modelName) {
	case "fable":
		return claudeTier{model: anthropic.ModelClaudeFable5_1, inputPrice: 10, outputPrice: 50, longInputPrice: 0, longOutputPrice: 0}
	case "opus":
		return claudeTier{model: anthropic.ModelClaudeOpus5_5, inputPrice: 4, outputPrice: 20, longInputPrice: 0, longOutputPrice: 0}
	case "sonnet":
		return claudeTier{model: anthropic.ModelClaudeSonnet5_5, inputPrice: 2, outputPrice: 10, longInputPrice: 0, longOutputPrice: 0}
	default:
		return claudeTier{model: anthropic.ModelClaudeHaiku5_5, inputPrice: 0.10, outputPrice: 0.50, longInputPrice: 0.50, longOutputPrice: 2.50}
	}
}

func (c *ClaudeClient) GetModelName() string {
	return string(c.tier.model)
}

func (c *ClaudeClient) calculateCost(inputTokens int64, outputTokens int64) float64 {
	if c.tier.longInputPrice > 0 && inputTokens > haikuLongPromptThreshold {
		return FormatCost(inputTokens, outputTokens, c.tier.longInputPrice, c.tier.longOutputPrice)
	}
	return FormatCost(inputTokens, outputTokens, c.tier.inputPrice, c.tier.outputPrice)
}

func (c *ClaudeClient) calculateBatchCost(inputTokens int64, outputTokens int64) float64 {
	return c.calculateCost(inputTokens, outputTokens) * batchCostMultiplier
}

func accumulateStream(client anthropic.Client, params anthropic.MessageNewParams) (anthropic.Message, error) {
	stream := client.Messages.NewStreaming(context.Background(), params)

	var message anthropic.Message
	for stream.Next() {
		if err := message.Accumulate(stream.Current()); err != nil {
			return message, err
		}
	}

	return message, stream.Err()
}

func (c *ClaudeClient) requireAPIKey() error {
	if c.apiKey == "" {
		return fmt.Errorf("Claude API key not configured. Please set your API key with: y config anthropic_api_key YOUR_API_KEY")
	}
	return nil
}

func (c *ClaudeClient) newClient() anthropic.Client {
	return anthropic.NewClient(option.WithAPIKey(c.apiKey))
}

func (c *ClaudeClient) buildParams(transaction logic.Transaction, systemPrompt string) (anthropic.MessageNewParams, int, error) {
	textBlocks, fileCount, err := BuildUserBlocks(transaction)
	if err != nil {
		return anthropic.MessageNewParams{}, 0, err
	}

	blocks := make([]anthropic.ContentBlockParamUnion, 0, len(textBlocks))
	for _, text := range textBlocks {
		blocks = append(blocks, anthropic.NewTextBlock(text))
	}
	messageParams := []anthropic.MessageParam{
		{
			Role:    anthropic.MessageParamRoleUser,
			Content: blocks,
		},
	}

	params := anthropic.MessageNewParams{
		Model:     c.tier.model,
		MaxTokens: int64(config.MaxTokens),
		Messages:  messageParams,
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	return params, fileCount, nil
}

func joinTextBlocks(message anthropic.Message) string {
	var responseText strings.Builder

	for _, block := range message.Content {
		responseText.WriteString(block.Text)
	}

	return responseText.String()
}

func (c *ClaudeClient) Call(transaction logic.Transaction, systemPrompt string) (string, error) {
	if err := c.requireAPIKey(); err != nil {
		return "", err
	}

	startTime := time.Now()

	client := c.newClient()

	params, fileCount, err := c.buildParams(transaction, systemPrompt)
	if err != nil {
		return "", err
	}

	fmt.Printf("Calling Claude with %d messages (%d files)\n", len(params.Messages), fileCount)

	message, err := accumulateStream(client, params)

	if err != nil {
		return "", fmt.Errorf("error calling Claude API: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Claude API call took %.2f seconds\n", duration.Seconds())
	fmt.Printf("Token usage - Input: %d, Output: %d\n",
		message.Usage.InputTokens,
		message.Usage.OutputTokens)

	responseText := joinTextBlocks(message)

	cost := c.calculateCost(message.Usage.InputTokens, message.Usage.OutputTokens)
	fmt.Printf("Cost: $%.6f\n", cost)

	if message.Usage.OutputTokens >= int64(config.MaxTokens) {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", config.MaxTokens)
	}

	trimmedResponse := strings.TrimSpace(responseText)

	return trimmedResponse, nil
}

func (c *ClaudeClient) toBatchRequestParams(params anthropic.MessageNewParams) anthropic.MessageBatchNewParamsRequestParams {
	return anthropic.MessageBatchNewParamsRequestParams{
		Model:     params.Model,
		MaxTokens: params.MaxTokens,
		Messages:  params.Messages,
		System:    params.System,
	}
}

func (c *ClaudeClient) SubmitBatch(transaction logic.Transaction, systemPrompt string) (string, error) {
	if err := c.requireAPIKey(); err != nil {
		return "", err
	}

	client := c.newClient()

	params, _, err := c.buildParams(transaction, systemPrompt)
	if err != nil {
		return "", err
	}

	batch, err := client.Messages.Batches.New(context.Background(), anthropic.MessageBatchNewParams{
		Requests: []anthropic.MessageBatchNewParamsRequest{
			{
				CustomID: batchCustomID,
				Params:   c.toBatchRequestParams(params),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("error submitting Claude batch: %w", err)
	}

	return batch.ID, nil
}

func (c *ClaudeClient) GetBatchStatus(batchID string) (BatchStatus, error) {
	if err := c.requireAPIKey(); err != nil {
		return BatchStatus{}, err
	}

	client := c.newClient()

	batch, err := client.Messages.Batches.Get(context.Background(), batchID, anthropic.MessageBatchGetParams{})
	if err != nil {
		return BatchStatus{}, fmt.Errorf("error fetching Claude batch: %w", err)
	}

	processingStatus := string(batch.ProcessingStatus)

	return BatchStatus{
		ProcessingStatus: processingStatus,
		Ended:            processingStatus == batchEndedStatus,
		Processing:       batch.RequestCounts.Processing,
		Succeeded:        batch.RequestCounts.Succeeded,
		Errored:          batch.RequestCounts.Errored,
		Canceled:         batch.RequestCounts.Canceled,
		Expired:          batch.RequestCounts.Expired,
	}, nil
}

func (c *ClaudeClient) GetBatchResult(batchID string) (BatchResult, error) {
	if err := c.requireAPIKey(); err != nil {
		return BatchResult{}, err
	}

	client := c.newClient()

	stream := client.Messages.Batches.ResultsStreaming(context.Background(), batchID, anthropic.MessageBatchResultsParams{})

	for stream.Next() {
		entry := stream.Current()
		if entry.CustomID == batchCustomID {
			return c.batchResultFromEntry(entry.Result), nil
		}
	}

	if err := stream.Err(); err != nil {
		return BatchResult{}, fmt.Errorf("error reading Claude batch results: %w", err)
	}

	return BatchResult{}, fmt.Errorf("no result found for request '%s' in batch %s", batchCustomID, batchID)
}

func (c *ClaudeClient) batchResultFromEntry(result anthropic.MessageBatchResultUnion) BatchResult {
	switch string(result.Type) {
	case "succeeded":
		return c.succeededBatchResult(result.Message)
	case "errored":
		return BatchResult{
			Failure: fmt.Sprintf("%s: %s", result.Error.Error.Type, result.Error.Error.Message),
		}
	case "canceled":
		return BatchResult{Failure: "canceled"}
	case "expired":
		return BatchResult{Failure: "expired"}
	default:
		return BatchResult{Failure: fmt.Sprintf("unknown result type: %s", result.Type)}
	}
}

func (c *ClaudeClient) succeededBatchResult(message anthropic.Message) BatchResult {
	return BatchResult{
		Text:         strings.TrimSpace(joinTextBlocks(message)),
		InputTokens:  message.Usage.InputTokens,
		OutputTokens: message.Usage.OutputTokens,
		Cost:         c.calculateBatchCost(message.Usage.InputTokens, message.Usage.OutputTokens),
	}
}