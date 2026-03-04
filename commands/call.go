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

func HandleActCommand(args []string, safe bool, think bool, cfg *config.Config, systemPrompt string) error {
	responseContent, err := HandleCall(args, think, cfg, systemPrompt, logic.TransactionTypeAct)
	if err != nil {
		return err
	}

	return processCodeBlocks(responseContent, safe)
}

func HandleVerbalCommand(args []string, think bool, cfg *config.Config, systemPrompt string, transactionType logic.TransactionType) error {
	responseContent, err := HandleCall(args, think, cfg, systemPrompt, transactionType)
	if err != nil {
		return err
	}

	fmt.Println("\n" + responseContent)
	return nil
}

func HandleGoCommand(think bool, cfg *config.Config, systemPrompt string) error {

	transactions, err := logic.LoadContextForMessageType(logic.TransactionTypeAct)
	if err != nil {
		fmt.Printf("Warning: could not load context: %v\n", err)
		transactions = []logic.Transaction{}
	}

	responseContent, err := callClaudeAPI(transactions, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	return processCodeBlocks(responseContent, false)
}

func HandleCall(args []string, think bool, cfg *config.Config, systemPrompt string, transactionType logic.TransactionType) (string, error) {
	prompt := strings.Join(args, " ")

	transactions, err := logic.LoadContextForMessageType(transactionType)
	if err != nil {
		fmt.Printf("Warning: could not load context: %v\n", err)
		transactions = []logic.Transaction{}
	}
	tx := transactions[len(transactions)-1]
	tx.Request = append(tx.Request, prompt)
	transactions[len(transactions)-1] = tx

	responseContent, err := callClaudeAPI(transactions, think, cfg, systemPrompt)
	if err != nil {
		return "", err
	}

	return responseContent, nil
}

func callClaudeAPI(transactions []logic.Transaction, think bool, cfg *config.Config, systemPrompt string) (string, error) {
	fmt.Printf("Sending request to Claude...\n")

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	transactions, err := client.Call(transactions, think, systemPrompt)

	done <- true
	close(done)

	if err != nil {
		return "", err
	}

	if err := logic.SaveContext(transactions); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	return transactions[len(transactions)-1].Response, nil
}

func processCodeBlocks(content string, safe bool) error {
	fmt.Println("Processing response...")
	var parseErrors []string
	codeblocks, text := logic.ParseCodeBlocks(content)
	for _, codeBlock := range codeblocks {
		err := codeBlock.Write(safe)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
		}
		if !codeBlock.Complete {
			fmt.Println("Warning: Incomplete code block.")
		}
	}

	if len(text) > 0 {
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
