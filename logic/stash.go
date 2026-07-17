// Stores and retrieves stashed prompts from the .yact/stash.txt file
package logic

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"yact/config"
)

const stashSeparator = "~~~~~~~~~~"

func PushStash() error {
	promptPath := config.GetProjectPromptPath()

	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("nothing to stash: prompt file does not exist")
		}
		return err
	}

	promptContent := strings.TrimSpace(string(promptData))
	if promptContent == "" {
		return errors.New("nothing to stash: prompt is empty")
	}

	stashPath := config.GetProjectStashPath()

	existingStash := ""
	stashData, err := os.ReadFile(stashPath)
	if err == nil {
		existingStash = string(stashData)
	} else if !os.IsNotExist(err) {
		return err
	}

	var newStash string
	if strings.TrimSpace(existingStash) != "" {
		newStash = promptContent + "\n" + stashSeparator + "\n" + existingStash
	} else {
		newStash = promptContent + "\n"
	}

	dir := filepath.Dir(stashPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(stashPath, []byte(newStash), 0644); err != nil {
		return err
	}

	emptyTransaction := Transaction{}
	return emptyTransaction.Save()
}

func PopStash() error {
	stashPath := config.GetProjectStashPath()

	stashData, err := os.ReadFile(stashPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("nothing to pop: stash is empty")
		}
		return err
	}

	stashContent := string(stashData)
	if strings.TrimSpace(stashContent) == "" {
		return errors.New("nothing to pop: stash is empty")
	}

	entries := strings.Split(stashContent, "\n"+stashSeparator+"\n")

	poppedPrompt := entries[0]
	remainingEntries := entries[1:]

	promptPath := config.GetProjectPromptPath()
	if err := os.WriteFile(promptPath, []byte(poppedPrompt), 0644); err != nil {
		return err
	}

	remainingStash := strings.Join(remainingEntries, "\n"+stashSeparator+"\n")

	dir := filepath.Dir(stashPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(stashPath, []byte(remainingStash), 0644)
}