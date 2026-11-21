package commands

import (
	"errors"
	"fmt"

	"yact/api"
	"yact/config"
	"yact/logic"
)

func HandleBatchCommand(cfg *config.Config) error {
	job, found, err := logic.ReadBatchJob()
	if err != nil {
		return err
	}
	if !found {
		fmt.Println("No batch job exists.")
		return nil
	}

	client := newClient(cfg)

	status, err := client.GetBatchStatus(job.ID)
	if err != nil {
		return err
	}

	if !status.Ended {
		printBatchStatus(status)
		return nil
	}

	result, err := client.GetBatchResult(job.ID)
	if err != nil {
		return err
	}

	if result.Failure != "" {
		if err := logic.DeleteBatchJobID(); err != nil {
			return err
		}
		return fmt.Errorf("batch failed: %s", result.Failure)
	}

	printBatchUsage(result)

	session := logic.Session{
		Linked:   job.Linked,
		Mode:     job.Mode,
		Response: result.Text,
	}

	// The batch result is paid for, so it is stored before the job id is deleted.
	// If storing fails, the job id stays and the result can be fetched again.
	if err := storeBuffer(session); err != nil {
		return err
	}

	deliverErr := deliverResponse(session)
	deleteErr := logic.DeleteBatchJobID()
	return errors.Join(deliverErr, deleteErr)
}

func printBatchStatus(status api.BatchStatus) {
	fmt.Printf("Batch status: %s\n", status.ProcessingStatus)
	fmt.Printf("Processing: %d\n", status.Processing)
	fmt.Printf("Succeeded: %d\n", status.Succeeded)
	fmt.Printf("Errored: %d\n", status.Errored)
	fmt.Printf("Canceled: %d\n", status.Canceled)
	fmt.Printf("Expired: %d\n", status.Expired)
}

func printBatchUsage(result api.BatchResult) {
	fmt.Printf("Input tokens: %d\n", result.InputTokens)
	fmt.Printf("Output tokens: %d\n", result.OutputTokens)
	fmt.Printf("Cost (50%% batch discount applied): $%.4f\n", result.Cost)
}
