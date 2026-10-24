package logic

import (
	"os"
	"path/filepath"
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
	return config.GetProjectYactDir()
}

func IndexFilePath() string {
	return filepath.Join(yactDir(), indexFileName)
}

func RemoveIndexFile() error {
	if err := os.Remove(IndexFilePath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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

func AddMessage(content string) (string, error) {
	return writeFixedFile(messageFileName, content)
}

func AddOutput(content string) (string, error) {
	return writeFixedFile(outputFileName, content)
}

func LoadRequestContent() (RequestContent, error) {
	messages, err := existingMessageFiles()
	if err != nil {
		return RequestContent{}, err
	}

	return RequestContent{
		Prompt:   "",
		Files:    []string{},
		Messages: messages,
	}, nil
}

func RemoveMessages() error {
	for _, name := range messageFileNames {
		if err := os.Remove(filepath.Join(yactDir(), name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
