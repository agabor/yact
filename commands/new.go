package commands

import (
	"fmt"
	"os"
	"yact/config"
	"yact/logic"
)

func HandleNewCommand() error {
	tx := logic.Transaction{}
	err := tx.Save()
	fmt.Println("New context created")
	if err != nil {
		return err
	}
	if err := logic.RemoveMessages(); err != nil {
		return err
	}
	return removePlanFile()
}

func removePlanFile() error {
	err := os.Remove(config.GetProjectPlanPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}