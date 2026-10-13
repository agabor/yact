package commands

import (
	"fmt"
	"strconv"
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

func applyMaxInputLinesOverride(cfg *config.Config, maxInputLinesOverride int) {
	if maxInputLinesOverride > 0 {
		cfg.MaxInputLines = maxInputLinesOverride
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

func printLinkWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
}

func printStepWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
}

func applyPromptLinks(transaction logic.Transaction, prompt string) logic.Transaction {
	if prompt == "" {
		return transaction
	}

	links := logic.ResolvePromptLinks(prompt)
	printLinkWarnings(links.Warnings)

	if !links.HasLinks {
		return transaction
	}

	linkedTransaction := transaction
	linkedTransaction.Context = append([]string{}, links.Paths...)
	return linkedTransaction
}

func resolvePlanPrompt(prompt string, writePlan bool) (string, bool, error) {
	if prompt != "" || writePlan {
		return "", false, nil
	}

	planText, planFound, err := logic.ReadPlan()
	if err != nil {
		return "", false, err
	}
	if !planFound {
		return "", false, fmt.Errorf("no plan found in %s, provide a prompt or create a plan with --plan", config.GetProjectPlanPath())
	}

	return planText, true, nil
}

func validateStepsOptions(stepsSelector string, writePlan bool, prompt string) error {
	if stepsSelector == "" {
		return nil
	}
	if writePlan {
		return fmt.Errorf("--steps and --plan cannot be combined")
	}
	if prompt != "" {
		return fmt.Errorf("--steps cannot be combined with a prompt argument")
	}
	_, err := logic.ParseStepSelector(stepsSelector)
	return err
}

func formatStepNumbers(steps []logic.PlanStep) string {
	texts := make([]string, 0, len(steps))
	for _, number := range logic.StepNumbers(steps) {
		texts = append(texts, strconv.Itoa(number))
	}
	return strings.Join(texts, ", ")
}

func selectPlanSteps(planText string, selector string) (string, []logic.PlanStep, bool, error) {
	parsedPlan := logic.ParsePlan(planText)
	if selector == "" {
		if parsedPlan.AllChecked() {
			fmt.Println("All plan steps are done.")
			return "", nil, true, nil
		}
		return planText, parsedPlan.Steps, false, nil
	}

	stepSelector, err := logic.ParseStepSelector(selector)
	if err != nil {
		return "", nil, false, err
	}

	selection, err := logic.SelectSteps(parsedPlan, stepSelector)
	if err != nil {
		return "", nil, false, err
	}

	printStepWarnings(selection.Warnings)

	if selection.AllDone {
		fmt.Println("All plan steps are done.")
		return "", nil, true, nil
	}

	fmt.Printf("Plan steps: %s\n", formatStepNumbers(selection.Steps))

	return logic.BuildStepsPrompt(parsedPlan, selection.Steps), selection.Steps, false, nil
}

func tickPlanSteps(planText string, steps []logic.PlanStep) error {
	if len(steps) == 0 {
		return nil
	}

	if err := logic.WritePlan(logic.MarkStepsDone(planText, steps)); err != nil {
		return err
	}

	fmt.Printf("Marked done: %s\n", formatStepNumbers(steps))
	return nil
}

func buildPlanModeTransaction(selectedPrompt string) logic.Transaction {
	planTransaction := logic.Transaction{Request: []string{selectedPrompt}}
	return applyPromptLinks(planTransaction, selectedPrompt)
}

func buildPromptModeTransaction(transaction logic.Transaction, prompt string) logic.Transaction {
	promptTransaction := transaction
	if prompt != "" {
		promptTransaction.Request = []string{prompt}
	}
	return promptTransaction
}

func HandleCommand(noWrite bool, noProgress bool, validateCode bool, writePlan bool, cfg *config.Config, systemPrompt string, modelOverride string, maxInputLinesOverride int, stepsSelector string, prompt string) error {
	if err := validateStepsOptions(stepsSelector, writePlan, prompt); err != nil {
		return err
	}

	applyModelOverride(cfg, modelOverride)
	applyMaxInputLinesOverride(cfg, maxInputLinesOverride)

	planPrompt, planUsed, err := resolvePlanPrompt(prompt, writePlan)
	if err != nil {
		return err
	}

	transaction := logic.Transaction{}
	var transactionForLLM logic.Transaction
	var sentSteps []logic.PlanStep

	if planUsed {
		selectedPrompt, selectedSteps, allDone, err := selectPlanSteps(planPrompt, stepsSelector)
		if err != nil {
			return err
		}
		if allDone {
			return nil
		}
		prompt = selectedPrompt
		sentSteps = selectedSteps
		transactionForLLM = buildPlanModeTransaction(selectedPrompt)
	} else {
		transaction, err = logic.LoadTransaction()
		if err != nil {
			return err
		}
		transactionForLLM = buildPromptModeTransaction(transaction, prompt)
	}

	totalLength := 0

	if prompt != "" {
		totalLength += len(strings.Split(prompt, "\n"))
	}

	for _, contextPath := range transactionForLLM.Context {
		codeFile, err := logic.ReadAsFile(contextPath)
		if err != nil {
			return err
		}
		totalLength += len(strings.Split(codeFile.Content, "\n"))
	}

	if totalLength > cfg.MaxInputLines {
		return fmt.Errorf("input exceeds maximum number allowed lines (%d), got %d. Increase max_input_lines to override.", cfg.MaxInputLines, totalLength)
	}

	fmt.Printf("Input lines: %d\n", totalLength)

	response, err := callLLM(transactionForLLM, noProgress, cfg, systemPrompt)
	if err != nil {
		return err
	}

	if writePlan {
		if err := logic.WritePlan(response); err != nil {
			return err
		}
		fmt.Printf("Plan written to %s\n", config.GetProjectPlanPath())
		return nil
	}

	if validateCode {
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
		fmt.Printf("Input lines: %d\n", totalLength)
		return err
	}

	if planUsed {
		return tickPlanSteps(planPrompt, sentSteps)
	}

	if err = transaction.Save(); err != nil {
		fmt.Printf("Input lines: %d\n", totalLength)
		return err
	}

	return nil
}

func newClient(cfg *config.Config) api.Client {
	var client api.Client

	modelType := strings.ToLower(cfg.ClaudeModel)
	if modelType == "qmax" || modelType == "qcoder" || modelType == "qnext" {
		client = &api.BedrockClient{}
	} else {
		client = &api.ClaudeClient{}
	}

	client.Init(cfg)

	return client
}

func callLLM(transaction logic.Transaction, noProgress bool, cfg *config.Config, systemPrompt string) (string, error) {
	client := newClient(cfg)

	if noProgress {
		return client.Call(transaction, cfg, systemPrompt)
	}

	fmt.Printf("Sending request...\n")
	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	response, err := client.Call(transaction, cfg, systemPrompt)

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