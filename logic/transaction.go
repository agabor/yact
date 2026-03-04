package logic

import "strings"

func LoadContextForMessageType(transactionType TransactionType) ([]Transaction, error) {
	transactions, err := LoadContext()
	if err != nil {
		return nil, err
	}

	newTransactions := make([]Transaction, 0)
	newTransaction := Transaction{Type: transactionType}

	for _, tx := range transactions {
		if tx.Type == transactionType {
			newTransactions = append(newTransactions, tx)
		} else {
			for _, file := range tx.Context {
				newTransaction.Context = append(newTransaction.Context, file)
			}
		}
	}

	if transactionType == TransactionTypeAct {
		lastPlan := strings.TrimSpace(getLastPlan(transactions))
		if lastPlan != "" {
			newTransaction.Request = []string{lastPlan}
		}
	}

	if transactionType == TransactionTypePlan {
		questions := getQuestions(transactions)
		if len(questions) > 0 {
			qaPrompt := "Questions and Answers\n" +
				"====================="
			for _, tx := range questions {
				qaPrompt = qaPrompt +
					"\nQuestion\n--------\n" + tx.Request[0] +
					"\nAnswer\n------\n" + tx.Response
			}
			newTransaction.Request = []string{qaPrompt}
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

func getQuestions(transactions []Transaction) []Transaction {
	var result []Transaction
	for _, tx := range transactions {
		if tx.Type == TransactionTypeQuestion {
			result = append(result, tx)
		}
	}
	return result
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
