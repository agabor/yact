package logic

import (
	"os"
	"strings"

	"yact/config"
)

func ReadPlan() (string, bool, error) {
	data, err := os.ReadFile(config.GetProjectPromptPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", false, nil
	}

	return content, true, nil
}

func WritePlan(content string) error {
	if err := os.MkdirAll(config.GetProjectYactDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(config.GetProjectPromptPath(), []byte(content), 0644)
}

func FormatPromptLink(path string) string {
	target := path
	if strings.Contains(path, " ") {
		target = "<" + path + ">"
	}
	return "[" + path + "](" + target + ")"
}

func HasPromptLink(promptText string, path string) bool {
	return strings.Contains(promptText, FormatPromptLink(path))
}

func readPromptFileIfExists() (string, error) {
	data, err := os.ReadFile(config.GetProjectPromptPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func withTrailingNewline(content string) string {
	if content == "" || strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}

func AppendPromptLink(path string) (bool, error) {
	if err := os.MkdirAll(config.GetProjectYactDir(), 0755); err != nil {
		return false, err
	}

	existingContent, err := readPromptFileIfExists()
	if err != nil {
		return false, err
	}

	if HasPromptLink(existingContent, path) {
		return false, nil
	}

	updatedContent := withTrailingNewline(existingContent) + FormatPromptLink(path) + "\n"
	if err := os.WriteFile(config.GetProjectPromptPath(), []byte(updatedContent), 0644); err != nil {
		return false, err
	}

	return true, nil
}