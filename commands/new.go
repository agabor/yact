package commands

import (
	"fmt"
	"yact/logic"
)

func HandleNewCommand() error {
	err := logic.SaveContext(logic.Transaction{})
	fmt.Println("New context created")
	return err
}
