package logic

import (
	"os"
	"path/filepath"
	"strings"
	"yact/config"
)

type Transaction struct {
	Request []string
	Context []string
}

func LoadTransaction() (Transaction, error) {
	contextPath := config.GetProjectContextPath()

	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Transaction{}, nil
		}
		return Transaction{}, err
	}

	transaction := Transaction{
		Request: []string{},
		Context: []string{},
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			transaction.Context = append(transaction.Context, trimmedLine)
		}
	}

	return transaction, nil
}

func (transaction *Transaction) Save() error {
	contextPath := config.GetProjectContextPath()

	dir := filepath.Dir(contextPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var builder strings.Builder

	for _, contextLine := range transaction.Context {
		builder.WriteString(contextLine)
		builder.WriteString("\n")
	}

	return os.WriteFile(contextPath, []byte(builder.String()), 0644)
}