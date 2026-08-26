package api

import (
	"context"
	"fmt"
	"strings"
	"time"
	"yact/config"
	"yact/logic"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type BedrockClient struct {
	region          string
	model           string
	maxOutputTokens int
	inputPrice      float64
	outputPrice     float64
}

func (c *BedrockClient) Init(cfg *config.Config) {
	c.region = cfg.AWSRegion
	c.model = cfg.BedrockModel
	c.maxOutputTokens = cfg.MaxTokens
	c.inputPrice = 0.22
	c.outputPrice = 0.95
}

func (c *BedrockClient) GetModelName() string {
	return c.model
}

func (c *BedrockClient) calculateCost(inputTokens int64, outputTokens int64) float64 {
	return FormatCost(inputTokens, outputTokens, c.inputPrice, c.outputPrice)
}

func (c *BedrockClient) buildMessages(transaction logic.Transaction) ([]types.Message, int, error) {
	blocks, fileCount, err := BuildUserBlocks(transaction)
	if err != nil {
		return nil, 0, err
	}

	contentBlocks := make([]types.ContentBlock, 0, len(blocks))
	for _, block := range blocks {
		contentBlocks = append(contentBlocks, &types.ContentBlockMemberText{Value: block})
	}

	messages := []types.Message{
		{
			Role:    types.ConversationRoleUser,
			Content: contentBlocks,
		},
	}

	return messages, fileCount, nil
}

func (c *BedrockClient) extractText(output types.ConverseOutput) string {
	messageOutput, ok := output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return ""
	}

	var responseText strings.Builder
	for _, block := range messageOutput.Value.Content {
		if textBlock, ok := block.(*types.ContentBlockMemberText); ok {
			responseText.WriteString(textBlock.Value)
		}
	}

	return responseText.String()
}

func (c *BedrockClient) Call(transaction logic.Transaction, think bool, systemPrompt string) (string, error) {
	ctx := context.Background()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.region))
	if err != nil {
		return "", fmt.Errorf("unable to load AWS configuration for region '%s'. Make sure AWS credentials and region are configured: %w", c.region, err)
	}

	client := bedrockruntime.NewFromConfig(awsCfg)

	messages, fileCount, err := c.buildMessages(transaction)
	if err != nil {
		return "", err
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:  aws.String(c.model),
		Messages: messages,
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens: aws.Int32(int32(c.maxOutputTokens)),
		},
	}

	if systemPrompt != "" {
		input.System = []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: systemPrompt},
		}
	}

	if think {
		fmt.Println("Note: extended thinking mode is not supported for this model, ignoring.")
	}

	fmt.Printf("Calling Bedrock with %d messages (%d files)\n", len(messages), fileCount)

	startTime := time.Now()

	resp, err := client.Converse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("error calling AWS Bedrock: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Bedrock API call took %.2f seconds\n", duration.Seconds())

	var inputTokens int64
	var outputTokens int64
	if resp.Usage != nil {
		if resp.Usage.InputTokens != nil {
			inputTokens = int64(*resp.Usage.InputTokens)
		}
		if resp.Usage.OutputTokens != nil {
			outputTokens = int64(*resp.Usage.OutputTokens)
		}
	}

	fmt.Printf("Token usage - Input: %d, Output: %d\n", inputTokens, outputTokens)

	cost := c.calculateCost(inputTokens, outputTokens)
	fmt.Printf("Cost: $%.6f\n", cost)

	if resp.StopReason == types.StopReasonMaxTokens {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", c.maxOutputTokens)
	}

	trimmedResponse := strings.TrimSpace(c.extractText(resp.Output))

	if err := logic.PutBuffer(trimmedResponse); err != nil {
		return "", fmt.Errorf("error writing buffer: %w", err)
	}

	return trimmedResponse, nil
}