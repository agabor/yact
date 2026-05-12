package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ClaudeModel = "haiku"
	MaxTokens   = 16000
	ThinkBudget = 8000
)

type Config struct {
	AnthropicAPIKey string `json:"anthropic_api_key"`
	ClaudeModel     string `json:"claude_model"`
	MaxTokens       int    `json:"max_tokens"`
	ThinkBudget     int    `json:"think_budget"`
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

func getPromptsDir() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "systemprompts"), nil
}

func getPromptFile(name string) (string, error) {
	dir, err := getPromptsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".txt"), nil
}

func DefaultConfig() *Config {
	return &Config{
		AnthropicAPIKey: "",
		ClaudeModel:     ClaudeModel,
		MaxTokens:       MaxTokens,
		ThinkBudget:     ThinkBudget,
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

func LoadPrompt(name string) (string, error) {
	promptFile, err := getPromptFile(name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(promptFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}