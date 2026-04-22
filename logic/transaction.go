package logic

import (
	"strings"
)

func LoadContextForPlan() (Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return Transaction{Type: TransactionTypePlan}, err
	}
	tx := CompactTransactions(transactions)
	tx.Type = TransactionTypePlan

	return tx, nil
}

func LoadContextForAct() (Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return Transaction{Type: TransactionTypeAct}, err
	}

	tx := CompactTransactions(transactions)
	tx.Type = TransactionTypeAct

	if TransactionTypePlan == getType(transactions) {
		lastPlan := strings.TrimSpace(getLastPlan(transactions))
		if lastPlan != "" {
			tx.Request = []string{lastPlan}
		}
	}

	return tx, nil
}

func LoadContextForQuestion() (Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return Transaction{Type: TransactionTypeQuestion}, err
	}
	tx := CompactTransactions(transactions)
	tx.Type = TransactionTypeQuestion
	return tx, nil
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
		return tx.Type
	}
	return TransactionTypeNone
}

func CompactTransactions(transactions []Transaction) Transaction {
	seenPaths := make(map[string]bool)

	newTransaction := Transaction{Type: TransactionTypeNone}

	for _, transaction := range transactions {
		for _, file := range transaction.Context {
			if seenPaths[file.Path()] {
				continue
			}
			file2, err := ReadAsFile(file.Path())
			if err != nil {
				continue
			}
			seenPaths[file.Path()] = true
			newTransaction.Context = append(newTransaction.Context, file2)
		}
		if transaction.Type == TransactionTypeAct {
			blocks, _ := ParseCodeBlocks(transaction.Response)
			for _, block := range blocks {
				if seenPaths[block.Path()] {
					continue
				}

				seenPaths[block.Path()] = true
				newTransaction.Context = append(newTransaction.Context, block)
			}
		}
	}

	return newTransaction
}

type TransactionType string

const (
	TransactionTypeNone     TransactionType = "None"
	TransactionTypeAct      TransactionType = "Act"
	TransactionTypePlan     TransactionType = "Plan"
	TransactionTypeQuestion TransactionType = "Question"
)

type Transaction struct {
	Type                      TransactionType
	Request                   []string
	Response                  string
	ResponseThinking          string
	ResponseThinkingSignature string
	Context                   []CodeFile
}
