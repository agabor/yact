package logic

import (
	"os"
	"path/filepath"
	"strings"
	"yact/config"
)

type Transaction struct {
	Request  []string
	Context  []string
	Messages []string
}

func messagesPath() string {
	return filepath.Join(filepath.Dir(config.GetProjectContextPath()), "messages.txt")
}

func readPathList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	paths := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			paths = append(paths, trimmedLine)
		}
	}

	return paths, nil
}

func writePathList(path string, lines []string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var builder strings.Builder

	for _, line := range lines {
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return os.WriteFile(path, []byte(builder.String()), 0644)
}

func LoadTransaction() (Transaction, error) {
	context, err := readPathList(config.GetProjectContextPath())
	if err != nil {
		return Transaction{}, err
	}

	messages, err := readPathList(messagesPath())
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		Request:  []string{},
		Context:  context,
		Messages: messages,
	}, nil
}

func (transaction *Transaction) Save() error {
	if err := writePathList(config.GetProjectContextPath(), transaction.Context); err != nil {
		return err
	}

	return writePathList(messagesPath(), transaction.Messages)
}