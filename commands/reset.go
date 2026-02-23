package commands

import (
	"fmt"
	"strings"
	"yact/logic"
)

func HandleResetCommand() error {
	transactions, err := logic.LoadContext()
	if err != nil {
		return err
	}

	seenPaths := make(map[string]bool)
	var reloadErrors []string

	newTransaction := logic.Transaction{Type: logic.TransactionTypeNone}

	for _, transaction := range transactions {
		for _, file := range transaction.Context {
			content, err := logic.ReadAsCodeBlock(file.Path)
			if err != nil {
				reloadErrors = append(reloadErrors, fmt.Sprintf("could not reload %s: %v", file.Path, err))
				continue
			}
			seenPaths[file.Path] = true
			newTransaction.Context = append(newTransaction.Context, logic.File{Path: file.Path, Content: content})
		}
		if transaction.Type == logic.TransactionTypeAct {
			for _, block := range logic.ParseCodeBlocks(transaction.Response) {
				if seenPaths[block.Path] {
					continue
				}

				seenPaths[block.Path] = true
				newTransaction.Context = append(newTransaction.Context, logic.File{Path: block.Path, Content: block.Content})
			}
		}
	}

	if err := logic.SaveContext([]logic.Transaction{newTransaction}); err != nil {
		return err
	}

	if len(reloadErrors) > 0 {
		return fmt.Errorf("reloaded context with errors: %s", strings.Join(reloadErrors, "; "))
	}
	fmt.Println("Context files reloaded")
	return nil
}
