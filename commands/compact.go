package commands

import (
	"fmt"
	"os"

	"yact/config"
	"yact/logic"
)

func HandleCompactCommand() error {
	promptPath := config.GetProjectPromptPath()

	data, err := os.ReadFile(string(promptPath))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Nothing to compact: %s does not exist\n", promptPath)
			return nil
		}
		return err
	}

	compacted := logic.CompactPromptLinks(string(data))
	if err := os.WriteFile(string(promptPath), []byte(compacted), 0644); err != nil {
		return err
	}

	fmt.Printf("Compacted: %s\n", promptPath)

	return nil
}
