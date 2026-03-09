package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"yact/logic"
)

func HandleReadCommand(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: y read <file> [<file2> ...]")
		return fmt.Errorf("missing file argument")
	}

	for _, pattern := range args {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("error matching pattern %s: %w", pattern, err)
		}

		if len(matches) == 0 {
			fmt.Printf("No files found matching pattern: %s\n", pattern)
			continue
		}

		for _, filePath := range matches {
			info, err := os.Stat(filePath)
			if err != nil {
				fmt.Printf("Error accessing %s: %v\n", filePath, err)
				continue
			}

			if info.IsDir() {
				fmt.Printf("Skipping directory: %s\n", filePath)
				continue
			}
			transactions, err := logic.LoadContext()
			if err != nil {
				return err
			}

			if hasMessageWithPath(transactions, filePath) {
				fmt.Printf("Skipping: %s\n", filePath)
				continue
			} else {
				fmt.Printf("Reading: %s\n", filePath)
			}

			file, err := logic.ReadAsFile(filePath)
			if err != nil {
				return err
			}
			if !AppendFileToLastTransaction(transactions, file) {
				transaction := logic.Transaction{Type: logic.TransactionTypeNone, Context: []logic.CodeFile{file}}
				transactions = append(transactions, transaction)
			}

			err2 := logic.SaveContext(transactions)
			if err2 != nil {
				return err2
			}
		}
	}

	return nil
}

func AppendFileToLastTransaction(transactions []logic.Transaction, file logic.CodeFile) bool {
	count := len(transactions)
	if count > 0 {
		transaction := transactions[count-1]
		if transaction.Type == logic.TransactionTypeNone {
			transaction.Context = append(transaction.Context, file)
			transactions[count-1] = transaction
			return true
		}
	}
	return false
}

func hasMessageWithPath(transactions []logic.Transaction, path string) bool {
	for _, transaction := range transactions {
		for _, file := range transaction.Context {
			if file.Path == path {
				return true
			}
		}
	}
	return false
}