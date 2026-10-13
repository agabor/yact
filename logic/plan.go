package logic

import (
	"os"
	"strings"

	"yact/config"
)

func ReadPlan() (string, bool, error) {
	data, err := os.ReadFile(config.GetProjectPlanPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", false, nil
	}

	return content, true, nil
}

func WritePlan(content string) error {
	if err := os.MkdirAll(config.GetProjectYactDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(config.GetProjectPlanPath(), []byte(content), 0644)
}