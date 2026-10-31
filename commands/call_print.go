package commands

import (
	"fmt"
	"time"

	"yact/api"
	"yact/config"
	"yact/logic"
)

func showProgress(done chan bool) {
	chars := []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
	idx := 0
	for {
		select {
		case <-done:
			fmt.Print("\r \r")
			return
		default:
			fmt.Printf("\r%c", chars[idx%len(chars)])
			idx++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func printWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
}

func printSentFiles(requestContent logic.RequestContent) {
	for _, messagePath := range requestContent.MessagePaths() {
		fmt.Printf("Sending: %s\n", messagePath)
	}
	for _, contextPath := range requestContent.Files {
		fmt.Printf("Sending: %s\n", contextPath)
	}
}

func printSentPrompt(prompt string) {
	if prompt == "" {
		return
	}
	fmt.Println("Sending prompt:")
	fmt.Println(prompt)
}

func printCallResult(result api.CallResult) {
	fmt.Printf("Claude API call took %.2f seconds\n", result.Duration.Seconds())
	fmt.Printf("Files sent: %d\n", result.FileCount)
	fmt.Printf("Token usage - Input: %d, Output: %d\n", result.InputTokens, result.OutputTokens)
	fmt.Printf("Cost: $%.6f\n", result.Cost)

	if result.Truncated {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", config.MaxTokens)
	}
}

func printApplyOutcome(path logic.ProjectPath, outcome logic.ApplyOutcome) {
	switch outcome {
	case logic.OutcomeWritten:
		fmt.Printf("Written: %s\n", path)
	case logic.OutcomeDeleted:
		fmt.Printf("Deleted: %s\n", path)
	case logic.OutcomeNothingToDelete:
		fmt.Printf("Not found (nothing to delete): %s\n", path)
	}
}