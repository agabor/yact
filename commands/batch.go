package commands

import (
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
		fmt.Printf("Batch failed: %s\n", result.Failure)
		return logic.DeleteBatchJobID()
	}

	printBatchUsage(result)

	outputErr := applyBatchOutput(job, result.Text)
	deleteErr := logic.DeleteBatchJobID()
	if outputErr != nil {
		return outputErr
	}
	return deleteErr
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

func applyBatchOutput(job logic.BatchJob, content string) error {
	if err := logic.PutBuffer(content); err != nil {
		return err
	}

	if job.WritePlan {
		if err := logic.WritePlan(content); err != nil {
			return err
		}
		fmt.Printf("Plan written to %s\n", config.GetProjectPlanPath())
		return nil
	}

	if job.ValidateCode {
		return validateCodeOnlyResponse(content)
	}

	if job.NoWrite {
		fmt.Println("\n" + content)
		return nil
	}

	fmt.Println("Response written to buffer. Run 'y commit' to apply it.")
	return nil
}
