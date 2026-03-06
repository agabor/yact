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

	newTransaction, err := logic.CompactTransactions(transactions)

	if err != nil {
		return err
	}

	if err := logic.SaveContext([]logic.Transaction{newTransaction}); err != nil {
		return err
	}

	fmt.Println("Context files reloaded")
	return nil
}
