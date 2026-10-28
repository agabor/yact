package logic

import (
	"encoding/json"
	"os"
	"strings"

	"yact/config"
)

type BatchJob struct {
	ID   string            `json:"id"`
	Mode config.OutputMode `json:"mode"`
}

func ReadBatchJob() (BatchJob, bool, error) {
	data, err := os.ReadFile(config.GetProjectBatchJobPath())
	if err != nil {
		if os.IsNotExist(err) {
			return BatchJob{}, false, nil
		}
		return BatchJob{}, false, err
	}

	var job BatchJob
	if err := json.Unmarshal(data, &job); err != nil {
		return BatchJob{ID: strings.TrimSpace(string(data))}, true, nil
	}

	return job, true, nil
}

func WriteBatchJob(job BatchJob) error {
	if err := os.MkdirAll(config.GetProjectYactDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.GetProjectBatchJobPath(), data, 0644)
}

func DeleteBatchJobID() error {
	err := os.Remove(config.GetProjectBatchJobPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}