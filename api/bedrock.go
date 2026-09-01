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
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const (
	defaultBedrockMaxModelID   = "qwen.qwen3-235b-a22b-2507-v1:0"
	defaultBedrockCoderModelID = "qwen.qwen3-coder-30b-a3b-v1:0"
	defaultBedrockNextModelID  = "qwen.qwen3-coder-next"
)

type BedrockClient struct {
	region          string
	model           string
	apiKey          string
	maxOutputTokens int
	inputPrice      float64
	outputPrice     float64
}

func (c *BedrockClient) Init(cfg *config.Config) {
	c.region = cfg.AWSRegion
	c.model = c.selectModel(cfg)
	c.apiKey = cfg.AWSAPIKey
	c.maxOutputTokens = cfg.MaxTokens
	c.inputPrice, c.outputPrice = c.selectPrices(cfg)
}

func isCoderModelSelected(modelName string) bool {
	return strings.EqualFold(strings.TrimSpace(modelName), "qcoder")
}

func isNextModelSelected(modelName string) bool {
	return strings.EqualFold(strings.TrimSpace(modelName), "qnext")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func resolveCoderModelID(cfg *config.Config) string {
	return firstNonEmpty(cfg.BedrockCoderModel, defaultBedrockCoderModelID)
}

func resolveNextModelID(cfg *config.Config) string {
	return firstNonEmpty(cfg.BedrockNextModel, defaultBedrockNextModelID)
}

func resolveMaxModelID(cfg *config.Config) string {
	configured := firstNonEmpty(cfg.BedrockModel, defaultBedrockMaxModelID)
	if strings.EqualFold(configured, resolveCoderModelID(cfg)) || strings.EqualFold(configured, resolveNextModelID(cfg)) {
		return defaultBedrockMaxModelID
	}
	return configured
}

func (c *BedrockClient) selectModel(cfg *config.Config) string {
	if isCoderModelSelected(cfg.ClaudeModel) {
		return resolveCoderModelID(cfg)
	}
	if isNextModelSelected(cfg.ClaudeModel) {
		return resolveNextModelID(cfg)
	}
	return resolveMaxModelID(cfg)
}

func (c *BedrockClient) selectPrices(cfg *config.Config) (float64, float64) {
	if isCoderModelSelected(cfg.ClaudeModel) {
		return 0.22, 0.95
	}
	if isNextModelSelected(cfg.ClaudeModel) {
		return 0.22, 0.95
	}
	return 0.22, 0.88
}

func (c *BedrockClient) GetModelName() string {
	return c.model
}

func (c *BedrockClient) calculateCost(inputTokens int64, outputTokens int64) float64 {
	return FormatCost(inputTokens, outputTokens, c.inputPrice, c.outputPrice)
}

func bearerAuthMiddleware(token string) func(*middleware.Stack) error {
	setAuthorizationHeader := middleware.FinalizeMiddlewareFunc(
		"YactBearerAuth",
		func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
			request, ok := in.Request.(*smithyhttp.Request)
			if !ok {
				return middleware.FinalizeOutput{}, middleware.Metadata{}, fmt.Errorf("unexpected request type %T", in.Request)
			}
			request.Header.Set("Authorization", "Bearer "+token)
			return next.HandleFinalize(ctx, in)
		},
	)

	return func(stack *middleware.Stack) error {
		stack.Finalize.Remove("Signing")
		return stack.Finalize.Add(setAuthorizationHeader, middleware.After)
	}
}

func (c *BedrockClient) newRuntimeClient(ctx context.Context) (*bedrockruntime.Client, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(c.region)}

	if c.apiKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS configuration for region '%s'. Make sure AWS credentials and region are configured, or set an API key with: y config aws_api_key YOUR_KEY: %w", c.region, err)
	}

	if c.apiKey == "" {
		return bedrockruntime.NewFromConfig(awsCfg), nil
	}

	return bedrockruntime.NewFromConfig(awsCfg, func(o *bedrockruntime.Options) {
		o.APIOptions = append(o.APIOptions, bearerAuthMiddleware(c.apiKey))
	}), nil
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

	client, err := c.newRuntimeClient(ctx)
	if err != nil {
		return "", err
	}

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
