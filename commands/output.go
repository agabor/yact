package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"yact/logic"
)

func HandleOutputCommand(args []string) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	content := string(input)
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("no output content provided")
	}

	outputPath, err := logic.AddOutput(content)
	if err != nil {
		return err
	}

	fmt.Println(outputPath)
	return nil
}