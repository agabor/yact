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
		printConfigValue("opus_model", cfg.OpusModel, false)
		printConfigValue("opus_input_price", fmt.Sprintf("%d", cfg.OpusInputPrice), false)
		printConfigValue("opus_output_price", fmt.Sprintf("%d", cfg.OpusOutputPrice), false)
		printConfigValue("opus_max_tokens", fmt.Sprintf("%d", cfg.OpusMaxTokens), false)
		printConfigValue("sonnet_model", cfg.SonnetModel, false)
		printConfigValue("sonnet_input_price", fmt.Sprintf("%d", cfg.SonnetInputPrice), false)
		printConfigValue("sonnet_output_price", fmt.Sprintf("%d", cfg.SonnetOutputPrice), false)
		printConfigValue("sonnet_max_tokens", fmt.Sprintf("%d", cfg.SonnetMaxTokens), false)
		printConfigValue("haiku_model", cfg.HaikuModel, false)
		printConfigValue("haiku_input_price", fmt.Sprintf("%d", cfg.HaikuInputPrice), false)
		printConfigValue("haiku_output_price", fmt.Sprintf("%d", cfg.HaikuOutputPrice), false)
		printConfigValue("haiku_max_tokens", fmt.Sprintf("%d", cfg.HaikuMaxTokens), false)
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
		case "opus_model":
			cfg.OpusModel = value
		case "opus_input_price":
			if err := setIntValue(&cfg.OpusInputPrice, value); err != nil {
				return err
			}
		case "opus_output_price":
			if err := setIntValue(&cfg.OpusOutputPrice, value); err != nil {
				return err
			}
		case "opus_max_tokens":
			if err := setIntValue(&cfg.OpusMaxTokens, value); err != nil {
				return err
			}
		case "sonnet_model":
			cfg.SonnetModel = value
		case "sonnet_input_price":
			if err := setIntValue(&cfg.SonnetInputPrice, value); err != nil {
				return err
			}
		case "sonnet_output_price":
			if err := setIntValue(&cfg.SonnetOutputPrice, value); err != nil {
				return err
			}
		case "sonnet_max_tokens":
			if err := setIntValue(&cfg.SonnetMaxTokens, value); err != nil {
				return err
			}
		case "haiku_model":
			cfg.HaikuModel = value
		case "haiku_input_price":
			if err := setIntValue(&cfg.HaikuInputPrice, value); err != nil {
				return err
			}
		case "haiku_output_price":
			if err := setIntValue(&cfg.HaikuOutputPrice, value); err != nil {
				return err
			}
		case "haiku_max_tokens":
			if err := setIntValue(&cfg.HaikuMaxTokens, value); err != nil {
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
