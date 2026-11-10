package logic

import (
	"os"
	"yact/config"
)

func ensureBufferDir() error {
	return os.MkdirAll(string(config.GetProjectTempDir()), 0755)
}

func PutBuffer(content string) error {
	if err := ensureBufferDir(); err != nil {
		return err
	}
	file, err := os.Create(string(BufferFilePath()))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func ReadBuffer() (string, error) {
	data, err := os.ReadFile(string(BufferFilePath()))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}
