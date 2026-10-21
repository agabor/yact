package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"yact/logic"
)

func HandleMessageCommand(args []string) error {
	if len(args) > 0 {
		fmt.Println("Usage: y message")
		return fmt.Errorf("unexpected arguments")
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	content := string(input)
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