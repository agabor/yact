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

func applyModelOverride(cfg *config.Config, modelOverride string) {
	if modelOverride != "" {
		cfg.ClaudeModel = modelOverride
	}
}

func HandleActCommand(think bool, noWrite bool, cfg *config.Config, systemPrompt string, modelOverride string, prompt string) error {
	applyModelOverride(cfg, modelOverride)

	if prompt != "" {
		if err := SetPrompt(prompt); err != nil {
			return err
		}
	}

	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	response, err := callClaudeAPI(transaction, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	if noWrite {
		fmt.Println("\n" + response)
		return nil
	}

	transaction, err = processResponse(transaction, response)
	if err != nil {
		return err
	}

	if err = transaction.Save(); err != nil {
		return err
	}

	return nil
}

func callClaudeAPI(transaction logic.Transaction, think bool, cfg *config.Config, systemPrompt string) (string, error) {

	fmt.Printf("Sending request to Claude...\n")

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	response, err := client.Call(transaction, think, systemPrompt)

	done <- true
	close(done)

	return response, err
}

func processResponse(transaction logic.Transaction, content string) (logic.Transaction, error) {
	var parseErrors []string
	elements := logic.ParseCodeFiles(content)

	seenPaths := make(map[string]bool)
	for _, filePath := range transaction.Context {
		seenPaths[filePath] = true
	}

	for _, element := range elements {
		switch v := element.(type) {
		case logic.CodeFile:
			err := v.Write()

			if err != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
			} else {
				if !seenPaths[v.Path] {
					seenPaths[v.Path] = true
					transaction.Context = append(transaction.Context, v.Path)
				}
			}
		case string:
			fmt.Println(v)
		}
	}

	if len(parseErrors) > 0 {
		return transaction, fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	return transaction, nil
}
