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

	var messages []api.Message

	for _, tx := range transactions {
		for _, file := range tx.Context {
			messages = append(messages, api.Message{Role: api.RoleTypeUser, Content: file.Content})
		}
		for _, prompt := range tx.Request {
			messages = append(messages, api.Message{Role: api.RoleTypeUser, Content: prompt})
		}
		if strings.TrimSpace(tx.Response) != "" {
			messages = append(messages, api.Message{Role: api.RoleTypeAssistant, Content: tx.Response, Thinking: tx.ResponseThinking, ThinkingSignature: tx.ResponseThinkingSignature})
		}
	}

	done := make(chan bool)
	go showProgress(done)

	response, err := client.Call(messages, think, systemPrompt)

	done <- true
	close(done)

	if err != nil {
		return "", err
	}

	if strings.TrimSpace(response.Content) == "" {
		return "", fmt.Errorf("error: empty response from Claude API")
	}

	transactions[len(transactions)-1].Response = response.Content
	transactions[len(transactions)-1].ResponseThinking = response.Thinking

	if err := logic.SaveContext(transactions); err != nil {
		fmt.Printf("Warning: could not save context: %v\n", err)
	}

	return response.Content, nil
}

func processCodeBlocks(content string, safe bool) error {
	fmt.Println("Processing response...")
	var parseErrors []string
	for _, codeBlock := range logic.ParseCodeBlocks(content) {
		err := codeBlock.Write(safe)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
		}
	}

	if len(parseErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	fmt.Println("Done!")
	return nil
}
