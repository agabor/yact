package commands

import (
	"fmt"
	"strings"

	"yact/config"
)

func HandleConfigCommand(args []string, cfg *config.Config) error {
	if len(args) == 0 {
		fmt.Println("Current configuration:")
		printConfigValue("anthropic_api_key", cfg.AnthropicAPIKey, true)
		printConfigValue("claude_model", cfg.ClaudeModel, false)
		printConfigValue("bedrock_model", cfg.BedrockModel, false)
		printConfigValue("bedrock_coder_model", cfg.BedrockCoderModel, false)
		printConfigValue("bedrock_next_model", cfg.BedrockNextModel, false)
		printConfigValue("aws_region", cfg.AWSRegion, false)
		printConfigValue("aws_api_key", cfg.AWSAPIKey, true)
		printConfigValue("max_tokens", fmt.Sprintf("%d", cfg.MaxTokens), false)
		printConfigValue("think_budget", fmt.Sprintf("%d", cfg.ThinkBudget), false)
		printConfigValue("max_input_lines", fmt.Sprintf("%d", cfg.MaxInputLines), false)
		return nil
	}

	if len(args) == 2 {
		key := args[0]
		value := args[1]

		if key == "ext" {
			return addExtension(value)
		}

		switch key {
		case "anthropic_api_key":
			cfg.AnthropicAPIKey = value
		case "claude_model":
			cfg.ClaudeModel = value
		case "bedrock_model":
			cfg.BedrockModel = value
		case "bedrock_coder_model":
			cfg.BedrockCoderModel = value
		case "bedrock_next_model":
			cfg.BedrockNextModel = value
		case "aws_region":
			cfg.AWSRegion = value
		case "aws_api_key":
			cfg.AWSAPIKey = value
		case "max_tokens":
			if err := setIntValue(&cfg.MaxTokens, value); err != nil {
				return err
			}
		case "think_budget":
			if err := setIntValue(&cfg.ThinkBudget, value); err != nil {
				return err
			}
		case "max_input_lines":
			if err := setIntValue(&cfg.MaxInputLines, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown config key '%s'", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("error saving config: %w", err)
		}

		if key == "anthropic_api_key" || key == "aws_api_key" {
			fmt.Printf("Set %s to %s\n", key, strings.Repeat("*", len(value)))
		} else {
			fmt.Printf("Set %s to %s\n", key, value)
		}
		return nil
	}

	fmt.Println("Usage:")
	fmt.Println("  y config                    # Show current config")
	fmt.Println("  y config <key> <value>      # Set config value")
	fmt.Println("  y config ext <extension>    # Add file extension to index")
	return nil
}

func addExtension(extension string) error {
	extension = strings.TrimPrefix(strings.TrimSpace(extension), ".")
	if extension == "" {
		return fmt.Errorf("extension cannot be empty")
	}

	extensions, err := config.LoadExtensions()
	if err != nil {
		return fmt.Errorf("error loading extensions: %w", err)
	}

	for _, existing := range extensions {
		if existing == extension {
			fmt.Printf("Extension '%s' already exists\n", extension)
			return nil
		}
	}

	extensions = append(extensions, extension)

	if err := config.SaveExtensions(extensions); err != nil {
		return fmt.Errorf("error saving extensions: %w", err)
	}

	fmt.Printf("Added extension '%s'\n", extension)
	return nil
}

func printConfigValue(key, value string, isSensitive bool) {
	if isSensitive && value != "" {
		fmt.Printf("  %s: %s\n", key, strings.Repeat("*", len(value)))
	} else {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

func setIntValue(target *int, value string) error {
	var intVal int
	_, err := fmt.Sscanf(value, "%d", &intVal)
	if err != nil {
		return fmt.Errorf("invalid integer value '%s'", value)
	}
	*target = intVal
	return nil
}