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

func HandleActCommand(think bool, noWrite bool, cfg *config.Config, systemPrompt string) error {
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

	codeFiles, transaction, err := processCodeBlocks(transaction, response)
	if err != nil {
		return err
	}

	if err = transaction.Save(); err != nil {
		return err
	}

	if len(codeFiles) > 0 {
		if err := updateIndexWithDescriptions(codeFiles); err != nil {
			fmt.Printf("Warning: Could not update index with descriptions: %v\n", err)
		}
	}

	return nil
}

func HandlePlanCommand(think bool, cfg *config.Config, systemPrompt string) error {
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

	fmt.Println("\n" + response)
	return nil
}

func HandleAskCommand(think bool, cfg *config.Config, systemPrompt string) error {
	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	response, err := callClaudeAPI(transaction, think, cfg, systemPrompt)

	fmt.Println("\n" + response)
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

func processCodeBlocks(transaction logic.Transaction, content string) ([]logic.CodeFile, logic.Transaction, error) {
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
		for _, text := range text {
			fmt.Println(text)
		}
		fmt.Println("")
	}

	if len(parseErrors) > 0 {
		return codeblocks, transaction, fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	fmt.Println("Done!")
	return codeblocks, transaction, nil
}

func updateIndexWithDescriptions(codeFiles []logic.CodeFile) error {
	indexFile := "yact-index.txt"

	indexedFiles, err := logic.LoadIndex(indexFile)
	if err != nil {
		return err
	}

	indexMap := make(map[string]*logic.FileEntry)
	for i := range indexedFiles {
		indexMap[indexedFiles[i].Path] = &indexedFiles[i]
	}

	updated := false
	for _, codeFile := range codeFiles {
		if codeFile.Description != "" {
			if entry, exists := indexMap[codeFile.Path]; exists {
				if entry.Description != codeFile.Description {
					entry.Description = codeFile.Description
					updated = true
				}
			}
		}
	}

	if updated {
		if err := logic.SaveIndex(indexFile, indexedFiles); err != nil {
			return err
		}
		fmt.Println("Index descriptions updated")
	}

	return nil
}