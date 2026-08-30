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

func validateCodeOnlyResponse(content string) error {
	elements, incomplete := logic.ParseCodeFilesDetailed(content)
	if incomplete {
		return fmt.Errorf("response contains an incomplete code block")
	}

	for _, element := range elements {
		if !element.IsCodeFile {
			if strings.TrimSpace(element.Text) != "" {
				return fmt.Errorf("response contains free text outside code blocks")
			}
		}
	}

	return nil
}

func HandleCommand(think bool, noWrite bool, quiet bool, codeOnly bool, noContext bool, noSavePrompt bool, cfg *config.Config, systemPrompt string, modelOverride string, prompt string) error {
	applyModelOverride(cfg, modelOverride)

	if prompt != "" && !noSavePrompt {
		if err := SetPrompt(prompt); err != nil {
			return err
		}
	}

	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	transactionForLLM := transaction
	if noContext {
		transactionForLLM.Context = []string{}
	}

	response, err := callLLM(transactionForLLM, think, quiet, cfg, systemPrompt)
	if err != nil {
		return err
	}

	if codeOnly {
		if err := validateCodeOnlyResponse(response); err != nil {
			return err
		}
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

	modelType := strings.ToLower(cfg.ClaudeModel)
	if modelType == "qmax" || modelType == "qcoder" {
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
		if element.IsCodeFile {
			v := element.CodeFile
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
		} else {
			if strings.TrimSpace(element.Text) != "" {
				fmt.Println(element.Text)
			}
		}
	}

	if len(parseErrors) > 0 {
		return transaction, fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	return transaction, nil
}