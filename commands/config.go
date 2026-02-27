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
		printConfigValue("max_tokens", fmt.Sprintf("%d", cfg.MaxTokens), false)
		printConfigValue("think_budget", fmt.Sprintf("%d", cfg.ThinkBudget), false)
		return nil
	}

	if len(args) == 2 {
		key := args[0]
		value := args[1]

		switch key {
		case "anthropic_api_key":
			cfg.AnthropicAPIKey = value
		case "claude_model":
			cfg.ClaudeModel = value
		case "max_tokens":
			if err := setIntValue(&cfg.MaxTokens, value); err != nil {
				return err
			}
		case "think_budget":
			if err := setIntValue(&cfg.ThinkBudget, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown config key '%s'", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("error saving config: %w", err)
		}

		if key == "anthropic_api_key" {
			fmt.Printf("Set %s to %s\n", key, strings.Repeat("*", len(value)))
		} else {
			fmt.Printf("Set %s to %s\n", key, value)
		}
		return nil
	}

	fmt.Println("Usage:")
	fmt.Println("  y config                    # Show current config")
	fmt.Println("  y config <key> <value>      # Set config value")
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
