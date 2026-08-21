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

func resolvePromptText(useBuffer bool, args []string) (string, error) {
	if useBuffer {
		return logic.ReadBuffer()
	}
	return args[0], nil
}

func HandlePromptCommand(args []string, useBuffer bool) error {
	prompt, err := resolvePromptText(useBuffer, args)
	if err != nil {
		return err
	}
	return SetPrompt(prompt)
}