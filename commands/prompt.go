// Handles setting the task prompt in .yact/prompt.txt while preserving the file context list
package commands

import (
	"yact/logic"
)

func SetPrompt(prompt string) error {
	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	transaction.Request = []string{prompt}

	return transaction.Save()
}

func HandlePromptCommand(args []string) error {
	return SetPrompt(args[0])
}