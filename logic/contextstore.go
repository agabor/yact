package logic

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const contextPath = ".yact"

func LoadContext() ([]Transaction, error) {
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

func SaveTransaction(transaction Transaction) error {
	return SaveContext([]Transaction{transaction})
}

func SaveContext(transactions []Transaction) error {
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
