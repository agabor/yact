package logic

import (
	"strings"
)

func LoadContextForPlan() ([]Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return nil, err
	}

	if TransactionTypeAct == getType(transactions) {
		tx := CompactTransactions(transactions)
		tx.Type = TransactionTypePlan
		return []Transaction{tx}, nil
	}

	return prepareNewTransaction(TransactionTypePlan, transactions)
}

func LoadContextForAct() ([]Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return nil, err
	}

	if TransactionTypePlan == getType(transactions) {
		tx := CompactTransactions(transactions)
		tx.Type = TransactionTypeAct
		lastPlan := strings.TrimSpace(getLastPlan(transactions))
		if lastPlan != "" {
			tx.Request = []string{lastPlan}
		}
		return []Transaction{tx}, nil
	}

	return prepareNewTransaction(TransactionTypeAct, transactions)
}

func prepareNewTransaction(t TransactionType, transactions []Transaction) ([]Transaction, error) {
	if len(transactions) > 0 && transactions[len(transactions)-1].Type == TransactionTypeNone {
		transactions[len(transactions)-1].Type = t
		return transactions, nil
	}

	return append(transactions, Transaction{Type: TransactionTypePlan}), nil
}

func LoadContextForQuestion() ([]Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return nil, err
	}
	tx := CompactTransactions(transactions)
	tx.Type = TransactionTypeQuestion
	return []Transaction{tx}, nil
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
			path := strings.TrimPrefix(strings.TrimSpace(file.Path), "./")
			if seenPaths[path] {
				continue
			}
			file2, err := ReadAsFile(path)
			if err != nil {
				continue
			}
			seenPaths[path] = true
			newTransaction.Context = append(newTransaction.Context, file2)
		}
		if transaction.Type == TransactionTypeAct {
			blocks, _ := ParseCodeBlocks(transaction.Response)
			for _, block := range blocks {
				block.Path = strings.TrimPrefix(strings.TrimSpace(block.Path), "./")
				if seenPaths[block.Path] {
					continue
				}

				seenPaths[block.Path] = true
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
