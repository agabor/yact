package commands

import (
	"fmt"
	"os"
	"yact/config"
	"yact/logic"
)

func HandleNewCommand() error {
	if err := logic.RemoveMessages(); err != nil {
		return err
	}
	if err := removePromptFile(); err != nil {
		return err
	}
	fmt.Println("New context created")
	return nil
}

func removePromptFile() error {
	err := os.Remove(config.GetProjectPromptPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
