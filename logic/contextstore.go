package logic

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const contextPath = ".yact"

func LoadContext() (Transaction, error) {
	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Transaction{}, nil
		}
		return Transaction{}, err
	}

	var transaction Transaction
	if err := json.Unmarshal(data, &transaction); err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}

func SaveContext(transaction Transaction) error {
	dir := filepath.Dir(contextPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(transaction, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(contextPath, data, 0644)
}
