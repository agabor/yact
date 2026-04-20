package commands

import (
	"fmt"

	"yact/logic"
)

func HandleContextCommand() error {
	transactions, err := logic.LoadContext()
	if err != nil {
		return err
	}

	if len(transactions) == 0 {
		fmt.Println("Context is empty")
		return nil
	}

	for i, transaction := range transactions {

		fmt.Printf("[%d] %s:", i, transaction.Type)
		fmt.Println()
		fmt.Print(" - User: ")
		fmt.Println()
		for j, file := range transaction.Context {
			fmt.Printf(" - [%d] %s", j, file.Path())
			fmt.Println()
		}
		for _, req := range transaction.Request {
			printTruncated(req)
		}
		printLine()
		fmt.Print(" - Assistant: ")
		fmt.Println()
		printTruncated(transaction.Response)
		printLine()
	}

	return nil
}

func printLine() {
	fmt.Println("--------------------------------------")
}

func printTruncated(truncatedContent string) {
	printLine()
	if len(truncatedContent) > 100 {
		truncatedContent = truncatedContent[:100] + "..."
	}
	fmt.Println(truncatedContent)
}
