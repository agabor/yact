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
	transactions, err := logic.LoadContextForMessageType(logic.TransactionTypeAct)
	if err != nil {
		return err
	}

	transactions, err = callClaudeAPI(prompt, transactions, think, cfg, systemPrompt)

	if err = logic.SaveContext(transactions); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	return processCodeBlocks(transactions[len(transactions)-1].Response)
}

func HandlePlanCommand(prompt string, think bool, cfg *config.Config, systemPrompt string) error {
	transactions, err := logic.LoadContextForMessageType(logic.TransactionTypePlan)
	if err != nil {
		return err
	}

	transactions, err = callClaudeAPI(prompt, transactions, think, cfg, systemPrompt)

	if err := logic.SaveContext(transactions); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	fmt.Println("\n" + transactions[len(transactions)-1].Response)
	return nil
}

func HandleAskCommand(prompt string, think bool, cfg *config.Config, systemPrompt string) error {
	transactions, err := logic.LoadQuestionsContext()
	if err != nil {
		return err
	}

	transactions, err = callClaudeAPI(prompt, transactions, think, cfg, systemPrompt)

	if err := logic.SaveQuestionsContext(transactions); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	fmt.Println("\n" + transactions[len(transactions)-1].Response)
	return nil
}

func appendPromptToTransactions(prompt string, transactions []logic.Transaction) []logic.Transaction {
	tx := transactions[len(transactions)-1]
	tx.Request = append(tx.Request, prompt)
	transactions[len(transactions)-1] = tx
	return transactions
}

func callClaudeAPI(prompt string, transactions []logic.Transaction, think bool, cfg *config.Config, systemPrompt string) ([]logic.Transaction, error) {

	if strings.TrimSpace(prompt) != "" {
		transactions = appendPromptToTransactions(prompt, transactions)
	}

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

	return transactions, err
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
		if !codeBlock.Complete {
			fmt.Println("Warning: Incomplete code block.")
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
