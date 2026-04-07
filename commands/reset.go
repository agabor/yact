package commands

import (
	"fmt"
	"yact/logic"
)

func HandleResetCommand() error {
	transactions, err := logic.LoadContext()
	if err != nil {
		return err
	}

	newTransaction := logic.CompactTransactions(transactions)

	if err := logic.SaveContext([]logic.Transaction{newTransaction}); err != nil {
		return err
	}

	fmt.Println("Context files reloaded")
	return nil
}
