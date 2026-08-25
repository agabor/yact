package logic

import (
	"os"
	"path/filepath"
)

const bufferFilePath = ".yact/buffer.txt"
const bufferDirPath = ".yact"

func ensureBufferDir() error {
	return os.MkdirAll(bufferDirPath, 0755)
}

func PutBuffer(content string) error {
	if err := ensureBufferDir(); err != nil {
		return err
	}
	file, err := os.Create(filepath.Clean(bufferFilePath))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func ReadBuffer() (string, error) {
	data, err := os.ReadFile(filepath.Clean(bufferFilePath))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}
