package logic

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func LoadContext() ([]Transaction, error) {
	return loadContextFromJson("context.json")
}

func SaveContext(transactions []Transaction) error {
	compact, err := CompactTransactions(transactions)
	if err != nil {
		return err
	}
	err = SaveQuestionsContext([]Transaction{compact})
	if err != nil {
		return err
	}
	return saveContextToJson("context.json", transactions)
}

func LoadQuestionsContext() ([]Transaction, error) {
	return loadContextFromJson("questions.json")
}

func SaveQuestionsContext(transactions []Transaction) error {
	return saveContextToJson("questions.json", transactions)
}

func getContextFilePath(jsonFileName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".yact", jsonFileName), nil
}

func loadContextFromJson(jsonFileName string) ([]Transaction, error) {
	contextPath, err := getContextFilePath(jsonFileName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Transaction{}, nil
		}
		return nil, err
	}

	var transactions []Transaction
	if err := json.Unmarshal(data, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func saveContextToJson(jsonFileName string, transactions []Transaction) error {
	contextPath, err := getContextFilePath(jsonFileName)
	if err != nil {
		return err
	}

	dir := filepath.Dir(contextPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(contextPath, data, 0644)
}
