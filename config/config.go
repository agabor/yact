package config

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ClaudeModel  = "haiku"
	BedrockModel = "qwen.qwen3-coder-30b-a3b-v1:0"
	AWSRegion    = "us-west-2"
	MaxTokens    = 16000
	ThinkBudget  = 8000
)

const promptDownloadBaseURL = "https://raw.githubusercontent.com/agabor/yact/refs/heads/main/systemprompts/"

type Config struct {
	AnthropicAPIKey string `json:"anthropic_api_key"`
	ClaudeModel     string `json:"claude_model"`
	BedrockModel    string `json:"bedrock_model"`
	AWSRegion       string `json:"aws_region"`
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

func GetProjectYactDir() string {
	return ".yact"
}

func GetProjectPromptPath() string {
	return filepath.Join(GetProjectYactDir(), "prompt.txt")
}

func GetProjectIndexPath() string {
	return filepath.Join(GetProjectYactDir(), "index.csv")
}

func GetProjectExtensionsPath() string {
	return filepath.Join(GetProjectYactDir(), "extensions.txt")
}

func GetProjectTagsPath() string {
	return filepath.Join(GetProjectYactDir(), "tags.csv")
}

func DefaultConfig() *Config {
	return &Config{
		AnthropicAPIKey: "",
		ClaudeModel:     ClaudeModel,
		BedrockModel:    BedrockModel,
		AWSRegion:       AWSRegion,
		MaxTokens:       MaxTokens,
		ThinkBudget:     ThinkBudget,
	}
}

func DefaultExtensions() []string {
	return []string{
		"go", "cs", "js", "jsx", "ts",
		"tsx", "py", "java", "rb", "php",
		"c", "cpp", "h", "hpp", "rs",
		"kt", "swift", "htm", "html",
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

	if cfg.BedrockModel == "" {
		cfg.BedrockModel = BedrockModel
	}

	if cfg.AWSRegion == "" {
		cfg.AWSRegion = AWSRegion
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

func DownloadPrompt(name string) error {
	promptFile, err := getPromptFile(name)
	if err != nil {
		return err
	}

	dir, err := getPromptsDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	url := promptDownloadBaseURL + name + ".txt"

	fmt.Printf("Downloading system prompt for '%s'...\n", name)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &promptDownloadError{name: name, statusCode: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := os.WriteFile(promptFile, data, 0644); err != nil {
		return err
	}

	fmt.Printf("System prompt for '%s' downloaded successfully.\n", name)

	return nil
}

type promptDownloadError struct {
	name       string
	statusCode int
}

func (e *promptDownloadError) Error() string {
	return "failed to download prompt '" + e.name + "': HTTP status " + http.StatusText(e.statusCode)
}

func ListPromptNames() ([]string, error) {
	dir, err := getPromptsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if strings.HasSuffix(fileName, ".txt") {
			names = append(names, strings.TrimSuffix(fileName, ".txt"))
		}
	}

	sort.Strings(names)

	return names, nil
}

func LoadExtensions() ([]string, error) {
	extensionsPath := GetProjectExtensionsPath()

	data, err := os.ReadFile(extensionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultExtensions := DefaultExtensions()
			if saveErr := SaveExtensions(defaultExtensions); saveErr != nil {
				return defaultExtensions, saveErr
			}
			return defaultExtensions, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var extensions []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			extensions = append(extensions, trimmed)
		}
	}

	if len(extensions) == 0 {
		return DefaultExtensions(), nil
	}

	return extensions, nil
}

func SaveExtensions(extensions []string) error {
	yactDir := GetProjectYactDir()
	if err := os.MkdirAll(yactDir, 0755); err != nil {
		return err
	}

	content := strings.Join(extensions, "\n") + "\n"
	return os.WriteFile(GetProjectExtensionsPath(), []byte(content), 0644)
}

func LoadTags() (map[string][]string, error) {
	tagsPath := GetProjectTagsPath()

	data, err := os.ReadFile(tagsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]string), nil
		}
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	tags := make(map[string][]string)
	for _, record := range records {
		if len(record) == 0 {
			continue
		}
		filePath := record[0]
		fileTags := record[1:]
		tags[filePath] = fileTags
	}

	return tags, nil
}

func SaveTags(tags map[string][]string) error {
	yactDir := GetProjectYactDir()
	if err := os.MkdirAll(yactDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(GetProjectTagsPath())
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var filePaths []string
	for filePath := range tags {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		record := append([]string{filePath}, tags[filePath]...)
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func GetFilesByTag(tagName string) ([]string, error) {
	tags, err := LoadTags()
	if err != nil {
		return nil, err
	}

	var files []string
	for filePath, fileTags := range tags {
		for _, tag := range fileTags {
			if tag == tagName {
				files = append(files, filePath)
				break
			}
		}
	}

	sort.Strings(files)

	return files, nil
}