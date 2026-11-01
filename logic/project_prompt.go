package logic

import (
	"os"

	"yact/config"
)

func WriteProjectPrompt(text string) error {
	return os.WriteFile(config.GetProjectPromptPath(), []byte(text), 0644)
}