// Stores and retrieves conversation context and prompts from .yact directory
package logic

import (
	"os"
	"path/filepath"
	"strings"
	"yact/config"
)

const separator = "=========="

type Transaction struct {
	Request []string
	Context []string
}

func LoadTransaction() (Transaction, error) {
	promptPath := config.GetProjectPromptPath()

	data, err := os.ReadFile(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Transaction{}, nil
		}
		return Transaction{}, err
	}

	content := string(data)
	parts := strings.Split(content, separator)

	transaction := Transaction{
		Request: []string{},
		Context: []string{},
	}

	if len(parts) > 0 {
		contextLines := strings.Split(strings.TrimSpace(parts[0]), "\n")
		for _, line := range contextLines {
			if strings.TrimSpace(line) != "" {
				transaction.Context = append(transaction.Context, line)
			}
		}
	}

	if len(parts) > 1 {
		requestContent := strings.TrimSpace(parts[1])
		if requestContent != "" {
			transaction.Request = []string{requestContent}
		}
	}

	return transaction, nil
}

func (transaction *Transaction) Save() error {
	promptPath := config.GetProjectPromptPath()

	dir := filepath.Dir(promptPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var builder strings.Builder

	for _, contextLine := range transaction.Context {
		builder.WriteString(contextLine)
		builder.WriteString("\n")
	}

	builder.WriteString(separator)
	builder.WriteString("\n")

	if len(transaction.Request) > 0 && strings.TrimSpace(transaction.Request[0]) != "" {
		builder.WriteString(transaction.Request[0])
		builder.WriteString("\n")
	}

	return os.WriteFile(promptPath, []byte(builder.String()), 0644)
}
