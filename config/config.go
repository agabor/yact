package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	ClaudeModel   = "haiku"
	MaxTokens     = 128000
	MaxInputLines = 2000
)

const promptDownloadBaseURL = "https://raw.githubusercontent.com/agabor/yact/refs/heads/main/systemprompts/"

type Config struct {
	AnthropicAPIKey string `json:"anthropic_api_key"`
	ClaudeModel     string `json:"claude_model"`
	MaxInputLines   int    `json:"max_input_lines"`
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

func getPromptFile(name PromptName) (string, error) {
	dir, err := getPromptsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, string(name)+".txt"), nil
}

func getIncludeFile(name string) (string, error) {
	dir, err := getPromptsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func GetProjectYactDir() string {
	return ".yact"
}

func GetProjectExtensionsPath() string {
	return filepath.Join(GetProjectYactDir(), "extensions.txt")
}

func GetProjectIgnorePath() string {
	return filepath.Join(GetProjectYactDir(), "ignore.txt")
}

func GetProjectPromptPath() string {
	return "yact-prompt.md"
}

func GetProjectBatchJobPath() string {
	return filepath.Join(GetProjectYactDir(), "batch_job_id.txt")
}

func DefaultConfig() *Config {
	return &Config{
		AnthropicAPIKey: "",
		ClaudeModel:     ClaudeModel,
		MaxInputLines:   MaxInputLines,
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

func DefaultIgnoredFolders() []string {
	return []string{"vendor", "node_modules"}
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

	if cfg.MaxInputLines == 0 {
		cfg.MaxInputLines = MaxInputLines
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

func parseFlags(content string) []string {
	if !strings.HasPrefix(content, "flags: ") {
		return []string{}
	}

	firstLine := strings.Split(content, "\n")[0]
	flagsStr := strings.TrimSpace(strings.TrimPrefix(firstLine, "flags: "))

	var flags []string
	runes := []rune(flagsStr)
	for index := 0; index < len(runes); index++ {
		ch := runes[index]
		if (ch == 'c' || ch == 'i') && index+1 < len(runes) && unicode.IsDigit(runes[index+1]) {
			flags = append(flags, string(runes[index:index+2]))
			index++
			continue
		}
		flags = append(flags, string(ch))
	}
	return flags
}

func readInclude(name string) (string, error) {
	if err := validateIncludeName(name); err != nil {
		return "", err
	}

	includeFile, err := getIncludeFile(name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(includeFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func LoadPromptWithFlags(name PromptName) (string, []string, error) {
	promptFile, err := getPromptFile(name)
	if err != nil {
		return "", nil, err
	}

	data, err := os.ReadFile(promptFile)
	if err != nil {
		return "", nil, err
	}

	content, err := ExpandIncludes(string(data), readInclude)
	if err != nil {
		return "", nil, fmt.Errorf("expanding includes for '%s': %w", name, err)
	}

	flags := parseFlags(content)
	return content, flags, nil
}

func fetchPromptFile(fileName string) ([]byte, error) {
	dir, err := getPromptsDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	url := promptDownloadBaseURL + fileName

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &promptDownloadError{name: fileName, statusCode: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	includeFile, err := getIncludeFile(fileName)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(includeFile, data, 0644); err != nil {
		return nil, err
	}

	return data, nil
}

func downloadPromptTree(fileName string, visited map[string]bool) error {
	if visited[fileName] {
		return nil
	}
	visited[fileName] = true

	data, err := fetchPromptFile(fileName)
	if err != nil {
		return err
	}

	for _, includeName := range ParseIncludeNames(string(data)) {
		if err := validateIncludeName(includeName); err != nil {
			return err
		}
		if err := downloadPromptTree(includeName, visited); err != nil {
			return err
		}
	}

	return nil
}

func DownloadPrompt(name PromptName) error {
	fmt.Printf("Downloading system prompt for '%s'...\n", name)

	if err := downloadPromptTree(string(name)+".txt", map[string]bool{}); err != nil {
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

func ParseIgnoredFolders(content string) ([]string, error) {
	var folders []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.ContainsAny(trimmed, "/\\") {
			return nil, fmt.Errorf("ignored folder %q must be a root folder name", trimmed)
		}
		folders = append(folders, trimmed)
	}
	return folders, nil
}

// An empty ignore file means no folder is ignored, unlike extensions where
// an empty file falls back to the defaults.
func LoadIgnoredFolders() ([]string, error) {
	data, err := os.ReadFile(GetProjectIgnorePath())
	if err != nil {
		if os.IsNotExist(err) {
			defaultFolders := DefaultIgnoredFolders()
			if saveErr := SaveIgnoredFolders(defaultFolders); saveErr != nil {
				return defaultFolders, saveErr
			}
			return defaultFolders, nil
		}
		return nil, err
	}

	return ParseIgnoredFolders(string(data))
}

func SaveIgnoredFolders(folders []string) error {
	yactDir := GetProjectYactDir()
	if err := os.MkdirAll(yactDir, 0755); err != nil {
		return err
	}

	content := strings.Join(folders, "\n") + "\n"
	return os.WriteFile(GetProjectIgnorePath(), []byte(content), 0644)
}
