package logic

import (
	"strings"
)

func LoadContextForPlan() (Transaction, error) {
	transaction, err := LoadContext()
	if err != nil {
		return Transaction{Type: TransactionTypePlan}, err
	}
	transaction.Type = TransactionTypePlan

	return transaction, nil
}

func LoadContextForAct() (Transaction, error) {
	transaction, err := LoadContext()
	if err != nil {
		return Transaction{Type: TransactionTypeAct}, err
	}

	transaction.Type = TransactionTypeAct

	if transaction.Type == TransactionTypePlan {
		lastPlan := strings.TrimSpace(transaction.Response)
		if lastPlan != "" {
			transaction.Request = []string{lastPlan}
		}
	}

	return transaction, nil
}

func LoadContextForQuestion() (Transaction, error) {
	transaction, err := LoadContext()
	if err != nil {
		return Transaction{Type: TransactionTypeQuestion}, err
	}
	transaction.Type = TransactionTypeQuestion
	return transaction, nil
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