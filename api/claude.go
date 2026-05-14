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
	case "opus":
		return anthropic.ModelClaudeOpus4_6
	case "sonnet":
		return anthropic.ModelClaudeSonnet4_6
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
	inputCost := (float64(inputTokens) / 1_000_000) * float64(c.inputPrice)
	outputCost := (float64(outputTokens) / 1_000_000) * float64(c.outputPrice)
	return inputCost + outputCost
}

func (c *ClaudeClient) Call(transaction logic.Transaction, think bool, systemPrompt string, index []logic.FileEntry) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Claude API key not configured. Please set your API key with: y config anthropic_api_key YOUR_API_KEY")
	}

	startTime := time.Now()

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	messageParams := make([]anthropic.MessageParam, 0)

	fileCount := 0
	blocks := make([]anthropic.ContentBlockParamUnion, 0)
	for _, filePath := range transaction.Context {
		fileContent, err := Serialize(filePath, index)
		if err != nil {
			return "", err
		}
		blocks = append(blocks, anthropic.NewTextBlock(fileContent))
		fileCount += 1
	}
	for _, prompt := range transaction.Request {
		blocks = append(blocks, anthropic.NewTextBlock(prompt))
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

	if think {
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

	return strings.TrimSpace(responseText.String()), nil
}
