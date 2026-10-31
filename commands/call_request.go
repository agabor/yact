package commands

import (
	"fmt"

	"yact/api"
	"yact/config"
	"yact/logic"
)

func checkNoPendingBatch() error {
	_, found, err := logic.ReadBatchJob()
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("a batch is already pending")
	}
	return nil
}

func submitBatch(requestContent logic.RequestContent, cfg *config.Config, systemPrompt string, mode config.OutputMode) error {
	client := newClient(cfg)

	batchID, err := client.SubmitBatch(requestContent, systemPrompt)
	if err != nil {
		return err
	}

	job := logic.BatchJob{
		ID:   batchID,
		Mode: mode,
	}
	if err := logic.WriteBatchJob(job); err != nil {
		return err
	}

	fmt.Printf("Batch submitted: %s\n", batchID)
	fmt.Println("Run `y batch` later to fetch the result.")
	return nil
}

func newClient(cfg *config.Config) api.Client {
	client := &api.ClaudeClient{}
	client.Init(cfg)
	return client
}

func callLLM(requestContent logic.RequestContent, noProgress bool, cfg *config.Config, systemPrompt string) (api.CallResult, error) {
	client := newClient(cfg)

	if noProgress {
		return client.Call(requestContent, systemPrompt)
	}

	fmt.Printf("Sending request...\n")
	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	result, err := client.Call(requestContent, systemPrompt)

	done <- true
	close(done)

	return result, err
}