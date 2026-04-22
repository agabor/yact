package logic

import (
	"os"
	"strings"
)

const contextPath = ".yact"
const separator = "=========="

func LoadContext() (Transaction, error) {
	data, err := os.ReadFile(contextPath)
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

func SaveContext(transaction Transaction) error {
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

	return os.WriteFile(contextPath, []byte(builder.String()), 0644)
}