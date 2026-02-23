package commands

import (
	"fmt"
	"strconv"
	"yact/logic"
)

func HandlePop(args []string) error {
	transactions, err := logic.LoadContext()
	if err != nil {
		return err
	}

	numToPop := 1
	if len(args) > 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid number: %s", args[0])
		}
		numToPop = num
	}

	if numToPop > len(transactions) {
		numToPop = len(transactions)
	}

	transactions = transactions[:len(transactions)-numToPop]

	if err := logic.SaveContext(transactions); err != nil {
		return err
	}

	fmt.Printf("Removed %d message(s)\n", numToPop)
	return nil
}
