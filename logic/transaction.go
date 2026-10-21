package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"yact/config"
)

type Transaction struct {
	Request  []string
	Context  []string
	Messages []string
}

type messageFile struct {
	path  string
	index int
}

func messagesDir() string {
	return filepath.Dir(config.GetProjectContextPath())
}

func parseMessageIndex(fileName string) (int, bool) {
	if !strings.HasPrefix(fileName, "message-") || !strings.HasSuffix(fileName, ".txt") {
		return 0, false
	}

	indexText := strings.TrimSuffix(strings.TrimPrefix(fileName, "message-"), ".txt")
	index, err := strconv.Atoi(indexText)
	if err != nil || index < 1 || strconv.Itoa(index) != indexText {
		return 0, false
	}

	return index, true
}

func listMessageFiles() ([]string, error) {
	entries, err := os.ReadDir(messagesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	messageFiles := []messageFile{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		index, isMessage := parseMessageIndex(entry.Name())
		if !isMessage {
			continue
		}

		messageFiles = append(messageFiles, messageFile{
			path:  filepath.Join(messagesDir(), entry.Name()),
			index: index,
		})
	}

	sort.Slice(messageFiles, func(i, j int) bool {
		return messageFiles[i].index < messageFiles[j].index
	})

	paths := make([]string, 0, len(messageFiles))
	for _, messageFile := range messageFiles {
		paths = append(paths, messageFile.path)
	}

	return paths, nil
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

func AddMessage(content string) (string, error) {
	if err := os.MkdirAll(messagesDir(), 0755); err != nil {
		return "", err
	}

	existingPaths, err := listMessageFiles()
	if err != nil {
		return "", err
	}

	nextIndex := 1
	if len(existingPaths) > 0 {
		lastIndex, _ := parseMessageIndex(filepath.Base(existingPaths[len(existingPaths)-1]))
		nextIndex = lastIndex + 1
	}

	messagePath := filepath.Join(messagesDir(), fmt.Sprintf("message-%d.txt", nextIndex))
	if err := os.WriteFile(messagePath, []byte(content), 0644); err != nil {
		return "", err
	}

	return messagePath, nil
}

func LoadTransaction() (Transaction, error) {
	context, err := readPathList(config.GetProjectContextPath())
	if err != nil {
		return Transaction{}, err
	}

	messages, err := listMessageFiles()
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
	return writePathList(config.GetProjectContextPath(), transaction.Context)
}

func RemoveMessages() error {
	messagePaths, err := listMessageFiles()
	if err != nil {
		return err
	}

	for _, messagePath := range messagePaths {
		if err := os.Remove(messagePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}