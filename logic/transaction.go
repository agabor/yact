package logic

import (
	"fmt"
	"strings"
)


func LoadContextForMessageType(transactionType TransactionType) ([]Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return nil, err
	}

	newTransactions := make([]Transaction, 0)
	newTransaction := Transaction{Type: transactionType}

	if transactionType != getType(transactions) {
		newTransaction, err = CompactTransactions(transactions)
		if err != nil {
			return nil, err
		}
	} else
		newTransaction = transactions
	}

	if transactionType == TransactionTypeAct {
		lastPlan := strings.TrimSpace(getLastPlan(transactions))
		if lastPlan != "" {
			newTransaction.Request = []string{lastPlan}
		}
	}

	return append(newTransactions, newTransaction), nil
}

func getLastPlan(transactions []Transaction) string {
	lastPlan := ""
	for _, tx := range transactions {
		if tx.Type == TransactionTypePlan {
			lastPlan = tx.Response
		}
	}
	return lastPlan
}

func getType(transactions []Transaction) TransactionType {
	for _, tx := range transactions {
		return tx.TransactionType
	}
	return TransactionTypeNone
}

func CompactTransactions(transactions []Transaction) (Transaction, error) {
	seenPaths := make(map[string]bool)
	var reloadErrors []string

	newTransaction := Transaction{Type: TransactionTypeNone}

	for _, transaction := range transactions {
		for _, file := range transaction.Context {
			path := strings.TrimPrefix(strings.TrimSpace(file.Path), "./")
			if seenPaths[path] {
				continue
			}
			content, err := ReadAsCodeBlock(path)
			if err != nil {
				reloadErrors = append(reloadErrors, fmt.Sprintf("could not reload %s: %v", path, err))
				continue
			}
			seenPaths[path] = true
			newTransaction.Context = append(newTransaction.Context, File{Path: path, Content: content})
		}
		if transaction.Type == TransactionTypeAct {
			blocks, _ := ParseCodeBlocks(transaction.Response)
			for _, block := range blocks {
				path := strings.TrimPrefix(strings.TrimSpace(block.Path), "./")
				if seenPaths[path] {
					continue
				}

				seenPaths[path] = true
				newTransaction.Context = append(newTransaction.Context, File{Path: path, Content: block.Content})
			}
		}
	}

	if len(reloadErrors) > 0 {
		return newTransaction, fmt.Errorf("reloaded context with errors: %s", strings.Join(reloadErrors, "; "))
	}
	return newTransaction, nil
}


type TransactionType string

const (
	TransactionTypeNone     TransactionType = "None"
	TransactionTypeQuestion TransactionType = "Question"
	TransactionTypeAct      TransactionType = "Act"
	TransactionTypePlan     TransactionType = "Plan"
)

type File struct {
	Path    string
	Content string
}

type Transaction struct {
	Type                      TransactionType
	Request                   []string
	Response                  string
	ResponseThinking          string
	ResponseThinkingSignature string
	Context                   []File
}
