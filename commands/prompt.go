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
