package commands

import (
	"fmt"
	"strings"
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

func HandleActCommand(prompt string, think bool, cfg *config.Config, systemPrompt string) error {
	transaction, err := logic.LoadContextForAct()
	if err != nil {
		return err
	}

	transaction, err = callClaudeAPI(prompt, transaction, think, cfg, systemPrompt)

	if err = logic.SaveTransaction(transaction); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	return processCodeBlocks(transaction.Response)
}

func HandlePlanCommand(prompt string, think bool, cfg *config.Config, systemPrompt string) error {
	transaction, err := logic.LoadContextForPlan()
	if err != nil {
		return err
	}

	transaction, err = callClaudeAPI(prompt, transaction, think, cfg, systemPrompt)

	if err := logic.SaveTransaction(transaction); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	fmt.Println("\n" + transaction.Response)
	return nil
}

func HandleAskCommand(prompt string, think bool, cfg *config.Config, systemPrompt string) error {
	transaction, err := logic.LoadContextForQuestion()
	if err != nil {
		return err
	}

	transaction, err = callClaudeAPI(prompt, transaction, think, cfg, systemPrompt)

	fmt.Println("\n" + transaction.Response)
	return nil
}

func callClaudeAPI(prompt string, transaction logic.Transaction, think bool, cfg *config.Config, systemPrompt string) (logic.Transaction, error) {

	if strings.TrimSpace(prompt) != "" {
		transaction.Request = append(transaction.Request, prompt)
	}

	fmt.Printf("Sending request to Claude...\n")

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	transaction, err := client.Call(transaction, think, systemPrompt)

	done <- true
	close(done)

	return transaction, err
}

func processCodeBlocks(content string) error {
	fmt.Println("Processing response...")
	var parseErrors []string
	codeblocks, text := logic.ParseCodeBlocks(content)
	for _, codeBlock := range codeblocks {
		err := codeBlock.Write()
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
		}
	}

	if len(text) > 0 && strings.TrimSpace(strings.Join(text, "")) != "" {
		fmt.Println("Text outside of code blocks:")
		for _, text := range text {
			fmt.Println(text)
		}
		fmt.Println("")
	}

	if len(parseErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	fmt.Println("Done!")
	return nil
}
