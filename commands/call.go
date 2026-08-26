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

func HandleActCommand(think bool, noWrite bool, quiet bool, cfg *config.Config, systemPrompt string, modelOverride string, prompt string) error {
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

	response, err := callLLM(transaction, think, quiet, cfg, systemPrompt)
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

func newClient(cfg *config.Config) api.Client {
	var client api.Client

	if strings.ToLower(cfg.ClaudeModel) == "qwen" {
		client = &api.BedrockClient{}
	} else {
		client = &api.ClaudeClient{}
	}

	client.Init(cfg)

	return client
}

func callLLM(transaction logic.Transaction, think bool, quiet bool, cfg *config.Config, systemPrompt string) (string, error) {
	client := newClient(cfg)

	if quiet {
		return client.Call(transaction, think, systemPrompt)
	}

	fmt.Printf("Sending request...\n")
	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	response, err := client.Call(transaction, think, systemPrompt)

	done <- true
	close(done)

	return response, err
}

func removeFromContext(context []string, path string) []string {
	result := make([]string, 0, len(context))
	for _, existingPath := range context {
		if existingPath != path {
			result = append(result, existingPath)
		}
	}
	return result
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
			if v.IsEmpty() {
				err := v.Delete()
				if err != nil {
					parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
				} else {
					seenPaths[v.Path] = false
					transaction.Context = removeFromContext(transaction.Context, v.Path)
				}
			} else {
				err := v.Write()

				if err != nil {
					parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
				} else {
					if !seenPaths[v.Path] {
						seenPaths[v.Path] = true
						transaction.Context = append(transaction.Context, v.Path)
					}
				}
			}
		case string:
			if strings.TrimSpace(v) != "" {
				fmt.Println(v)
			}
		}
	}

	if len(parseErrors) > 0 {
		return transaction, fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	return transaction, nil
}