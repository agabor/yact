package api

import (
	"context"
	"fmt"
	"strings"
	"time"
	"yact/config"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeClient struct {
	apiKey          string
	model           string
	maxOutputTokens int
	inputPrice      int
	outputPrice     int
}

func (c *ClaudeClient) Init(cfg *config.Config) {
	c.apiKey = cfg.AnthropicAPIKey
	c.model = c.selectModel(cfg)
	c.maxOutputTokens = c.selectMaxTokens(cfg)
	c.inputPrice = c.selectInputPrice(cfg)
	c.outputPrice = c.selectOutputPrice(cfg)
}

func (c *ClaudeClient) selectModel(cfg *config.Config) string {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "opus":
		return cfg.OpusModel
	case "sonnet":
		return cfg.SonnetModel
	case "haiku":
		return cfg.HaikuModel
	default:
		return cfg.HaikuModel
	}
}

func (c *ClaudeClient) selectMaxTokens(cfg *config.Config) int {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "opus":
		return cfg.OpusMaxTokens
	case "sonnet":
		return cfg.SonnetMaxTokens
	case "haiku":
		return cfg.HaikuMaxTokens
	default:
		return cfg.HaikuMaxTokens
	}
}

func (c *ClaudeClient) selectInputPrice(cfg *config.Config) int {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "opus":
		return cfg.OpusInputPrice
	case "sonnet":
		return cfg.SonnetInputPrice
	case "haiku":
		return cfg.HaikuInputPrice
	default:
		return cfg.HaikuInputPrice
	}
}

func (c *ClaudeClient) selectOutputPrice(cfg *config.Config) int {
	switch strings.ToLower(cfg.ClaudeModel) {
	case "opus":
		return cfg.OpusOutputPrice
	case "sonnet":
		return cfg.SonnetOutputPrice
	case "haiku":
		return cfg.HaikuOutputPrice
	default:
		return cfg.HaikuOutputPrice
	}
}

func (c *ClaudeClient) GetModelName() string {
	return c.model
}

func (c *ClaudeClient) calculateCost(inputTokens int64, outputTokens int64) float64 {
	inputCost := (float64(inputTokens) / 1_000_000) * float64(c.inputPrice)
	outputCost := (float64(outputTokens) / 1_000_000) * float64(c.outputPrice)
	return inputCost + outputCost
}

func (c *ClaudeClient) Call(messages []Message, systemPrompt string) (Message, error) {
	if c.apiKey == "" {
		return Message{}, fmt.Errorf("Claude API key not configured. Please set your API key with: y config anthropic_api_key YOUR_API_KEY")
	}

	startTime := time.Now()

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	messageParams := make([]anthropic.MessageParam, len(messages))
	for i, msg := range messages {
		if msg.Role == RoleTypeAssistant {
			messageParams[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
		} else {
			messageParams[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(c.maxOutputTokens)),
		Messages:  anthropic.F(messageParams),
	}

	if systemPrompt != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		})
	}

	fmt.Printf("Calling Claude with %d messages\n", len(messages))

	message, err := client.Messages.New(context.Background(), params)

	if err != nil {
		return Message{}, fmt.Errorf("error calling Claude API: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Claude API call took %.2f seconds\n", duration.Seconds())
	fmt.Printf("Token usage - Input: %d, Output: %d\n",
		message.Usage.InputTokens,
		message.Usage.OutputTokens)

	cost := c.calculateCost(message.Usage.InputTokens, message.Usage.OutputTokens)
	fmt.Printf("Cost: $%.6f\n", cost)

	if message.Usage.OutputTokens >= int64(c.maxOutputTokens) {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", c.maxOutputTokens)
	}

	var responseText string
	for _, block := range message.Content {
		responseText += block.Text
	}

	return Message{
		Role:    RoleTypeAssistant,
		Content: responseText,
	}, nil
}
