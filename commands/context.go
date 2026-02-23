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

		count, _ := fmt.Printf("[%d] %s", i, transaction.Type)
		for _, req := range transaction.Request {
			fmt.Print(" - User: ")
			printTruncated(req)
		}
		fmt.Print(strings.Repeat(" ", count))
		fmt.Print(" - Assistant: ")
		printTruncated(transaction.Response)

		for j, file := range transaction.Context {
			fmt.Printf(" - [%d] %s", j, file.Path)
		}
	}

	return nil
}

func printTruncated(truncatedContent string) {
	if len(truncatedContent) > 200 {
		truncatedContent = truncatedContent[:200] + "..."
	}
	truncatedContent = strings.ReplaceAll(truncatedContent, "\n", " ")
	fmt.Printf(" - %s", truncatedContent)
	fmt.Println()
}
