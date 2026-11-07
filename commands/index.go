package commands

import (
	"fmt"

	"yact/config"
	"yact/logic"
)

func HandleIndexCommand(args []string) error {
	setting, err := config.ParseIndexSetting(args[0])
	if err != nil {
		return err
	}

	if err := applyIndexOverride(&setting); err != nil {
		return err
	}

	if setting == config.IndexNone {
		fmt.Printf("Index file removed: %s\n", logic.IndexFilePath())
		return nil
	}
	fmt.Printf("Index written to %s\n", logic.IndexFilePath())
	return nil
}
