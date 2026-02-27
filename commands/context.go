package commands

import (
	"fmt"
	"strings"

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
		for _, req := range transaction.Request {
			printTruncated(req)
		}
		fmt.Print(" - Assistant: ")
		printTruncated(transaction.Response)

		for j, file := range transaction.Context {
			fmt.Printf(" - [%d] %s", j, file.Path)
			fmt.Println()
		}
	}

	return nil
}

func printTruncated(truncatedContent string) {
	if len(truncatedContent) > 200 {
		truncatedContent = truncatedContent[:200] + "..."
	}
	truncatedContent = strings.ReplaceAll(truncatedContent, "\n", " ")
	fmt.Println(truncatedContent)
}
