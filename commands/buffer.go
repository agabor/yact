package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"yact/config"
)

func getBufferFilePath() string {
	return filepath.Join(config.GetProjectYactDir(), "buffer.txt")
}

func SaveBuffer(content string) error {
	yactDir := config.GetProjectYactDir()
	if err := os.MkdirAll(yactDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(getBufferFilePath(), []byte(content), 0644)
}

func GetBuffer() (string, error) {
	path := getBufferFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func HandleBufferCommand() error {
	content, err := GetBuffer()
	if err != nil {
		return err
	}
	fmt.Print(content)
	return nil
}
