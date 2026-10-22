package commands

import (
	"fmt"
	"strings"

	"yact/logic"
)

func HandleMessageCommand(args []string) error {
	content := args[0]
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("no message content provided")
	}

	messagePath, err := logic.AddMessage(content)
	if err != nil {
		return err
	}

	fmt.Println(messagePath)
	return nil
}