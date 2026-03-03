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
			path := strings.TrimPrefix(strings.TrimSpace(file.Path), "./")
			if seenPaths[path] {
				continue
			}
			content, err := logic.ReadAsCodeBlock(path)
			if err != nil {
				reloadErrors = append(reloadErrors, fmt.Sprintf("could not reload %s: %v", path, err))
				continue
			}
			seenPaths[path] = true
			newTransaction.Context = append(newTransaction.Context, logic.File{Path: path, Content: content})
		}
		if transaction.Type == logic.TransactionTypeAct {
			blocks, _ := logic.ParseCodeBlocks(transaction.Response)
			for _, block := range blocks {
				path := strings.TrimPrefix(strings.TrimSpace(block.Path), "./")
				if seenPaths[path] {
					continue
				}

				seenPaths[path] = true
				newTransaction.Context = append(newTransaction.Context, logic.File{Path: path, Content: block.Content})
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
