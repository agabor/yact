package logic

import (
	"os"
	"path/filepath"
	"strings"
	"yact/config"
)

type RequestContent struct {
	Prompt   string
	Files    []string
	Messages []string
}

const (
	messageFileName = "message.txt"
	indexFileName   = "index.txt"
	outputFileName  = "output.txt"
)

var messageFileNames = []string{messageFileName, indexFileName, outputFileName}

func yactDir() string {
	return filepath.Dir(config.GetProjectContextPath())
}

func IndexFilePath() string {
	return filepath.Join(yactDir(), indexFileName)
}

func writeFixedFile(name string, content string) (string, error) {
	if err := os.MkdirAll(yactDir(), 0755); err != nil {
		return "", err
	}

	path := filepath.Join(yactDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func existingMessageFiles() ([]string, error) {
	paths := []string{}
	for _, name := range messageFileNames {
		path := filepath.Join(yactDir(), name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		paths = append(paths, path)
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
	return writeFixedFile(messageFileName, content)
}

func AddOutput(content string) (string, error) {
	return writeFixedFile(outputFileName, content)
}

func LoadRequestContent() (RequestContent, error) {
	context, err := readPathList(config.GetProjectContextPath())
	if err != nil {
		return RequestContent{}, err
	}

	messages, err := existingMessageFiles()
	if err != nil {
		return RequestContent{}, err
	}

	return RequestContent{
		Prompt:   "",
		Files:    context,
		Messages: messages,
	}, nil
}

func (requestContent *RequestContent) Save() error {
	return writePathList(config.GetProjectContextPath(), requestContent.Files)
}

func RemoveMessages() error {
	for _, name := range messageFileNames {
		if err := os.Remove(filepath.Join(yactDir(), name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
