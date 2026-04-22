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
	lines := strings.Split(strings.TrimSpace(content), "\n")

	transaction := Transaction{
		Request: []string{},
		Context: []string{},
	}

	separatorIndex := -1
	for i, line := range lines {
		if line == separator {
			separatorIndex = i
			break
		}
	}

	if separatorIndex == -1 {
		return transaction, nil
	}

	for i := 0; i < separatorIndex; i++ {
		if strings.TrimSpace(lines[i]) != "" {
			transaction.Context = append(transaction.Context, lines[i])
		}
	}

	if separatorIndex+1 < len(lines) && strings.TrimSpace(lines[separatorIndex+1]) != "" {
		transaction.Request = append(transaction.Request, lines[separatorIndex+1])
	}

	return transaction, nil
}

func SaveContext(transaction Transaction) error {
	var builder strings.Builder

	for _, contextPath := range transaction.Context {
		builder.WriteString(contextPath)
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