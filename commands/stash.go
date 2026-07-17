// Handles stashing and restoring the current task prompt
package commands

import (
	"errors"
	"fmt"

	"yact/logic"
)

func HandleStashCommand(args []string) error {
	if len(args) == 0 {
		if err := logic.PushStash(); err != nil {
			return err
		}
		fmt.Println("Prompt stashed.")
		return nil
	}

	if len(args) == 1 && args[0] == "pop" {
		if err := logic.PopStash(); err != nil {
			return err
		}
		fmt.Println("Prompt restored from stash.")
		return nil
	}

	return errors.New("invalid usage: expected 'y stash' or 'y stash pop'")
}