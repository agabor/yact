package commands

import (
	"fmt"
	"yact/logic"
)

func HandleNewCommand() error {
	tx := logic.Transaction{}
	err := tx.Save()
	fmt.Println("New context created")
	return err
}
