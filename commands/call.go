// Handles act, plan, ask, and bash commands for code generation, planning, question-answering, and shell command operations
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

func HandleActCommand(think bool, noWrite bool, cfg *config.Config, systemPrompt string, modelOverride string) error {
	applyModelOverride(cfg, modelOverride)

	if err := logic.ClearBuffer(); err != nil {
		return err
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
		if err := logic.AppendToBuffer(response); err != nil {
			return err
		}
		fmt.Println("\n" + response)
		return nil
	}

	transaction, err = processCodeBlocks(transaction, response)
	if err != nil {
		return err
	}

	if err = transaction.Save(); err != nil {
		return err
	}

	return nil
}

func HandlePlanCommand(think bool, cfg *config.Config, systemPrompt string, modelOverride string) error {
	applyModelOverride(cfg, modelOverride)

	if err := logic.ClearBuffer(); err != nil {
		return err
	}

	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	response, err := callClaudeAPI(transaction, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	transaction.Request = []string{response}

	if err := transaction.Save(); err != nil {
		return err
	}

	if err := logic.AppendToBuffer(response); err != nil {
		return err
	}

	fmt.Println("\n" + response)
	return nil
}

func HandleAskCommand(think bool, cfg *config.Config, systemPrompt string, modelOverride string) error {
	applyModelOverride(cfg, modelOverride)

	if err := logic.ClearBuffer(); err != nil {
		return err
	}

	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	response, err := callClaudeAPI(transaction, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	if err := logic.AppendToBuffer(response); err != nil {
		return err
	}

	fmt.Println("\n" + response)
	return nil
}

func HandleBashCommand(think bool, cfg *config.Config, systemPrompt string, modelOverride string) error {
	applyModelOverride(cfg, modelOverride)

	if err := logic.ClearBuffer(); err != nil {
		return err
	}

	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	response, err := callClaudeAPI(transaction, think, cfg, systemPrompt)
	if err != nil {
		return err
	}

	if err := logic.AppendToBuffer(response); err != nil {
		return err
	}

	fmt.Println("\n" + response)
	return nil
}

func callClaudeAPI(transaction logic.Transaction, think bool, cfg *config.Config, systemPrompt string) (string, error) {

	fmt.Printf("Sending request to Claude...\n")

	var client api.Client
	client = &api.ClaudeClient{}
	client.Init(cfg)

	fmt.Printf("Model: %s\n", client.GetModelName())

	indexedFiles, err := logic.LoadIndex()
	if err != nil {
		return "", fmt.Errorf("error loading index: %w", err)
	}

	done := make(chan bool)
	go showProgress(done)

	response, err := client.Call(transaction, think, systemPrompt, indexedFiles)

	done <- true
	close(done)

	return response, err
}

func processCodeBlocks(transaction logic.Transaction, content string) (logic.Transaction, error) {
	var parseErrors []string
	codeblocks, text := logic.ParseCodeFiles(content)

	seenPaths := make(map[string]bool)
	for _, filePath := range transaction.Context {
		seenPaths[filePath] = true
	}

	for _, codeBlock := range codeblocks {
		err := codeBlock.Write()

		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
		} else {
			if !seenPaths[codeBlock.Path] {
				seenPaths[codeBlock.Path] = true
				transaction.Context = append(transaction.Context, codeBlock.Path)
			}
		}
	}

	if len(text) > 0 && strings.TrimSpace(strings.Join(text, "")) != "" {
		fmt.Println("Text outside of code blocks:")
		for _, textLine := range text {
			fmt.Println(textLine)
		}
		fmt.Println("")

		bufferContent := "Text outside of code blocks:\n" + strings.Join(text, "\n") + "\n"
		if err := logic.AppendToBuffer(bufferContent); err != nil {
			return transaction, err
		}
	}

	if len(parseErrors) > 0 {
		errorContent := "Errors processing code blocks:\n" + strings.Join(parseErrors, "\n") + "\n"
		if err := logic.AppendToBuffer(errorContent); err != nil {
			return transaction, err
		}
		return transaction, fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	fmt.Println("Done!")
	return transaction, nil
}