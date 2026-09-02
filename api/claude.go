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

type ClaudeClient struct {
	apiKey          string
	model           anthropic.Model
	maxOutputTokens int
	thinkBudget     int
	inputPrice      int
	outputPrice     int
}

func (c *ClaudeClient) Init(cfg *config.Config) {
	c.apiKey = cfg.AnthropicAPIKey
	c.model = c.selectModel(cfg)
	c.maxOutputTokens = cfg.MaxTokens
	c.thinkBudget = cfg.ThinkBudget
	c.inputPrice = c.selectInputPrice(cfg)
	c.outputPrice = c.selectOutputPrice(cfg)
}

func (c *ClaudeClient) selectModel(cfg *config.Config) anthropic.Model {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "fable":
		return anthropic.ModelClaudeFable5
	case "opus":
		return anthropic.ModelClaudeOpus5
	case "sonnet":
		return anthropic.ModelClaudeSonnet5
	default:
		return anthropic.ModelClaudeHaiku4_5
	}
}

func (c *ClaudeClient) selectInputPrice(cfg *config.Config) int {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "opus":
		return 5
	case "sonnet":
		return 3
	default:
		return 1
	}
}

func (c *ClaudeClient) selectOutputPrice(cfg *config.Config) int {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "opus":
		return 25
	case "sonnet":
		return 15
	default:
		return 5
	}
}

func (c *ClaudeClient) GetModelName() string {
	return string(c.model)
}

func (c *ClaudeClient) calculateCost(inputTokens int64, outputTokens int64) float64 {
	return FormatCost(inputTokens, outputTokens, float64(c.inputPrice), float64(c.outputPrice))
}

func (c *ClaudeClient) Call(transaction logic.Transaction, cfg *config.Config, systemPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Claude API key not configured. Please set your API key with: y config anthropic_api_key YOUR_API_KEY")
	}

	startTime := time.Now()

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	messageParams := make([]anthropic.MessageParam, 0)

	textBlocks, fileCount, err := BuildUserBlocks(transaction)
	if err != nil {
		return "", err
	}

	blocks := make([]anthropic.ContentBlockParamUnion, 0, len(textBlocks))
	for _, text := range textBlocks {
		blocks = append(blocks, anthropic.NewTextBlock(text))
	}
	messageParams = append(messageParams, anthropic.MessageParam{
		Role:    anthropic.MessageParamRoleUser,
		Content: blocks,
	})
	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: int64(c.maxOutputTokens),
		Messages:  messageParams,
	}

	if c.thinkBudget > 0 {
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(int64(c.thinkBudget))
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	fmt.Printf("Calling Claude with %d messages (%d files)\n", len(messageParams), fileCount)

	message, err := client.Messages.New(context.Background(), params)

	if err != nil {
		return "", fmt.Errorf("error calling Claude API: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Claude API call took %.2f seconds\n", duration.Seconds())
	fmt.Printf("Token usage - Input: %d, Output: %d\n",
		message.Usage.InputTokens,
		message.Usage.OutputTokens)

	var responseText strings.Builder
	var thinkingText string

	for _, block := range message.Content {
		if block.Type == "thinking" {
			thinkingText = block.Thinking
		} else {
			responseText.WriteString(block.Text)
		}
	}

	if thinkingText != "" {
		fmt.Println("\n=== Claude's Thinking ===")
		fmt.Println(thinkingText)
		fmt.Println("===")
		fmt.Println("")
	}

	cost := c.calculateCost(message.Usage.InputTokens, message.Usage.OutputTokens)
	fmt.Printf("Cost: $%.6f\n", cost)

	if message.Usage.OutputTokens >= int64(c.maxOutputTokens) {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", c.maxOutputTokens)
	}

	trimmedResponse := strings.TrimSpace(responseText.String())

	if err := logic.PutBuffer(trimmedResponse); err != nil {
		return "", fmt.Errorf("error writing buffer: %w", err)
	}

	return trimmedResponse, nil
}