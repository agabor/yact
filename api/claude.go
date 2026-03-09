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

func (c *ClaudeClient) Call(transactions []logic.Transaction, think bool, systemPrompt string) ([]logic.Transaction, error) {
	if c.apiKey == "" {
		return transactions, fmt.Errorf("Claude API key not configured. Please set your API key with: y config anthropic_api_key YOUR_API_KEY")
	}

	startTime := time.Now()

	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))

	messageParams := make([]anthropic.MessageParam, 0)

	for _, tx := range transactions {
		blocks := make([]anthropic.ContentBlockParamUnion, 0)
		for _, file := range tx.Context {
			blocks = append(blocks, anthropic.NewTextBlock(file.Serialize()))
		}
		for _, prompt := range tx.Request {
			blocks = append(blocks, anthropic.NewTextBlock(prompt))
		}
		messageParams = append(messageParams, anthropic.MessageParam{
			Role:    anthropic.MessageParamRoleUser,
			Content: blocks,
		})
		blocks = make([]anthropic.ContentBlockParamUnion, 0)
		if strings.TrimSpace(tx.Response) != "" {
			if strings.TrimSpace(tx.ResponseThinkingSignature) != "" && strings.TrimSpace(tx.ResponseThinking) != "" {
				blocks = append(blocks, anthropic.NewThinkingBlock(tx.ResponseThinkingSignature, tx.ResponseThinking))
			}
			blocks = append(blocks, anthropic.NewTextBlock(tx.Response))
			messageParams = append(messageParams, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleAssistant,
				Content: blocks,
			})
		}
	}
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

	fmt.Printf("Calling Claude with %d messages\n", len(messageParams))

	message, err := client.Messages.New(context.Background(), params)

	if err != nil {
		return transactions, fmt.Errorf("error calling Claude API: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Claude API call took %.2f seconds\n", duration.Seconds())
	fmt.Printf("Token usage - Input: %d, Output: %d\n",
		message.Usage.InputTokens,
		message.Usage.OutputTokens)

	var responseText string
	var thinkingText string
	var thinkingSignature string

	for _, block := range message.Content {
		if block.Type == "thinking" {
			thinkingText = block.Thinking
			thinkingSignature = block.Signature
		} else {
			responseText += block.Text
		}
	}

	if thinkingText != "" {
		fmt.Println("\n=== Claude's Thinking ===")
		fmt.Println(thinkingText)
		fmt.Println("===\n")
	}

	cost := c.calculateCost(message.Usage.InputTokens, message.Usage.OutputTokens)
	fmt.Printf("Cost: $%.6f\n", cost)

	if message.Usage.OutputTokens >= int64(c.maxOutputTokens) {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", c.maxOutputTokens)
	}

	tx := transactions[len(transactions)-1]
	tx.Response = responseText
	tx.ResponseThinking = thinkingText
	tx.ResponseThinkingSignature = thinkingSignature
	transactions[len(transactions)-1] = tx

	if strings.TrimSpace(responseText) == "" {
		return transactions, fmt.Errorf("error: empty response from Claude API")
	}

	return transactions, nil
}
