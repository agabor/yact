package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ClaudeModel       = "haiku"
	OpusModel         = "claude-opus-4-6"
	OpusInputPrice    = 5
	OpusOutputPrice   = 25
	OpusMaxTokens     = 128000
	SonnetModel       = "claude-sonnet-4-6"
	SonnetInputPrice  = 3
	SonnetOutputPrice = 15
	SonnetMaxTokens   = 64000
	HaikuModel        = "claude-haiku-4-5"
	HaikuInputPrice   = 1
	HaikuOutputPrice  = 5
	HaikuMaxTokens    = 64000
)

type Config struct {
	AnthropicAPIKey   string `json:"anthropic_api_key"`
	ClaudeModel       string `json:"claude_model"`
	OpusModel         string `json:"opus_model"`
	OpusInputPrice    int    `json:"opus_input_price"`
	OpusOutputPrice   int    `json:"opus_output_price"`
	OpusMaxTokens     int    `json:"opus_max_tokens"`
	SonnetModel       string `json:"sonnet_model"`
	SonnetInputPrice  int    `json:"sonnet_input_price"`
	SonnetOutputPrice int    `json:"sonnet_output_price"`
	SonnetMaxTokens   int    `json:"sonnet_max_tokens"`
	HaikuModel        string `json:"haiku_model"`
	HaikuInputPrice   int    `json:"haiku_input_price"`
	HaikuOutputPrice  int    `json:"haiku_output_price"`
	HaikuMaxTokens    int    `json:"haiku_max_tokens"`
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".yact"), nil
}

func getConfigFile() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config"), nil
}
func DefaultConfig() *Config {
	return &Config{
		AnthropicAPIKey:   "",
		ClaudeModel:       ClaudeModel,
		OpusModel:         OpusModel,
		OpusInputPrice:    OpusInputPrice,
		OpusOutputPrice:   OpusOutputPrice,
		OpusMaxTokens:     OpusMaxTokens,
		SonnetModel:       SonnetModel,
		SonnetInputPrice:  SonnetInputPrice,
		SonnetOutputPrice: SonnetOutputPrice,
		SonnetMaxTokens:   SonnetMaxTokens,
		HaikuModel:        HaikuModel,
		HaikuInputPrice:   HaikuInputPrice,
		HaikuOutputPrice:  HaikuOutputPrice,
		HaikuMaxTokens:    HaikuMaxTokens,
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	configFile, err := getConfigFile()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (c *Config) Save() error {
	configFile, err := getConfigFile()
	if err != nil {
		return err
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
