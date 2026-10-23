package logic

import (
	"os"
	"strings"

	"yact/config"
)

func ReadBatchJobID() (string, bool, error) {
	data, err := os.ReadFile(config.GetProjectBatchJobPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	batchJobID := strings.TrimSpace(string(data))

	return batchJobID, true, nil
}

func WriteBatchJobID(id string) error {
	if err := os.MkdirAll(config.GetProjectYactDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(config.GetProjectBatchJobPath(), []byte(id), 0644)
}

func DeleteBatchJobID() error {
	err := os.Remove(config.GetProjectBatchJobPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}