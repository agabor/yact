package commands

import (
	"fmt"
	"os"
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

func applyIndexOverride(flags config.Flaglist) error {
	if !flags.HasIndex() {
		return logic.RemoveIndexFile()
	}
	visibility, err := indexLevelToVisibility(flags.IndexLevel)
	if err != nil {
		return err
	}
	return WriteIndex(visibility)
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

func printWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
}

func applyPromptLinks(requestContent logic.RequestContent) (logic.RequestContent, error) {
	if requestContent.Prompt == "" {
		return requestContent, nil
	}

	links, err := logic.ResolvePromptLinks(requestContent.Prompt)
	if err != nil {
		return requestContent, err
	}

	requestContent.Files = append([]string{}, links.Paths...)
	return requestContent, nil
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

func selectPlanSteps(planText string, selector string) (string, []logic.PlanStep, bool, error) {
	parsedPlan := logic.ParsePlan(planText)
	if selector == "" {
		if len(parsedPlan.Steps) == 0 {
			return planText, parsedPlan.Steps, false, nil
		}
		if parsedPlan.AllChecked() {
			fmt.Println("All plan steps are done.")
			return "", nil, true, nil
		}
		remainingSteps := parsedPlan.UncheckedSteps()
		fmt.Printf("Plan steps: %s\n", logic.FormatNumbers(logic.StepNumbers(remainingSteps)))
		return logic.BuildStepsPrompt(parsedPlan, remainingSteps), remainingSteps, false, nil
	}

	stepSelector, err := logic.ParseStepSelector(selector)
	if err != nil {
		return "", nil, false, err
	}

	selection, err := logic.SelectSteps(parsedPlan, stepSelector)
	if err != nil {
		return "", nil, false, err
	}

	printWarnings(selection.Warnings)

	if selection.AllDone {
		fmt.Println("All plan steps are done.")
		return "", nil, true, nil
	}

	fmt.Printf("Plan steps: %s\n", logic.FormatNumbers(logic.StepNumbers(selection.Steps)))

	return logic.BuildStepsPrompt(parsedPlan, selection.Steps), selection.Steps, false, nil
}

func tickPlanSteps(planText string, steps []logic.PlanStep) error {
	if len(steps) == 0 {
		return nil
	}

	if err := logic.WritePlan(logic.MarkStepsDone(planText, steps)); err != nil {
		return err
	}

	fmt.Printf("Marked done: %s\n", logic.FormatNumbers(logic.StepNumbers(steps)))
	return nil
}

func combinePromptText(promptFileText string, cliPrompt string) string {
	fileText := strings.TrimSpace(promptFileText)
	cliText := strings.TrimSpace(cliPrompt)
	if fileText == "" {
		return cliText
	}
	if cliText == "" {
		return fileText
	}
	return fileText + "\n\n" + cliText
}

func checkNoPendingBatch() error {
	_, found, err := logic.ReadBatchJob()
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("a batch is already pending")
	}
	return nil
}

func submitBatch(requestContent logic.RequestContent, cfg *config.Config, systemPrompt string, flags config.Flaglist) error {
	client := newClient(cfg)

	batchID, err := client.SubmitBatch(requestContent, systemPrompt)
	if err != nil {
		return err
	}

	job := logic.BatchJob{
		ID:           batchID,
		NoWrite:      flags.NoWrite,
		ValidateCode: flags.ValidateCode,
		WritePlan:    flags.WritePlan,
	}
	if err := logic.WriteBatchJob(job); err != nil {
		return err
	}

	fmt.Printf("Batch submitted: %s\n", batchID)
	fmt.Println("Run `y batch` later to fetch the result.")
	return nil
}

func countInputLines(requestContent logic.RequestContent) (int, error) {
	totalLength := 0

	if requestContent.Prompt != "" {
		totalLength += len(strings.Split(requestContent.Prompt, "\n"))
	}

	for _, messagePath := range requestContent.MessagePaths() {
		messageContent, err := os.ReadFile(messagePath)
		if err != nil {
			return 0, err
		}
		totalLength += len(strings.Split(string(messageContent), "\n"))
	}

	for _, contextPath := range requestContent.Files {
		codeFile, err := logic.ReadAsFile(contextPath)
		if err != nil {
			return 0, err
		}
		totalLength += len(strings.Split(codeFile.Content, "\n"))
	}

	return totalLength, nil
}

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, prompt string) error {
	if flags.Batch {
		if err := checkNoPendingBatch(); err != nil {
			return err
		}
	}

	if err := validateStepsOptions(flags.Steps, flags.WritePlan, prompt); err != nil {
		return err
	}

	if err := applyIndexOverride(flags); err != nil {
		return err
	}

	applyModelOverride(cfg, flags.ModelOverride())
	applyMaxInputLinesOverride(cfg, flags.MaxInputLines)

	var err error
	planPrompt := ""
	planUsed := false

	if prompt == "" {
		var planFound bool
		planPrompt, planFound, err = logic.ReadPlan()
		if err != nil {
			return err
		}
		if !planFound {
			return fmt.Errorf("no plan found in %s, provide a prompt or create a plan with --plan", config.GetProjectPromptPath())
		}
		planUsed = true
	}

	requestContentForLLM, err := logic.LoadRequestContent()
	if err != nil {
		return err
	}
	var sentSteps []logic.PlanStep

	if planUsed {
		selectedPrompt, selectedSteps, allDone, err := selectPlanSteps(planPrompt, flags.Steps)
		if err != nil {
			return err
		}
		if allDone {
			return nil
		}
		sentSteps = selectedSteps
		requestContentForLLM.Prompt = selectedPrompt
	} else {
		promptFileText, _, err := logic.ReadPlan()
		if err != nil {
			return err
		}

		requestContentForLLM.Prompt = combinePromptText(promptFileText, prompt)
	}
	requestContentForLLM, err = applyPromptLinks(requestContentForLLM)
	if err != nil {
		return err
	}

	totalLength, err := countInputLines(requestContentForLLM)
	if err != nil {
		return err
	}

	if totalLength > cfg.MaxInputLines {
		return fmt.Errorf("input exceeds maximum number allowed lines (%d), got %d. Increase max_input_lines to override.", cfg.MaxInputLines, totalLength)
	}

	fmt.Printf("Input lines: %d\n", totalLength)

	if flags.Batch {
		return submitBatch(requestContentForLLM, cfg, systemPrompt, flags)
	}

	response, err := callLLM(requestContentForLLM, flags.NoProgress, cfg, systemPrompt)
	if err != nil {
		return err
	}

	if err := storeBuffer(response); err != nil {
		return err
	}

	if flags.WritePlan {
		if err := logic.WritePlan(response); err != nil {
			return err
		}
		fmt.Printf("Plan written to %s\n", config.GetProjectPromptPath())
		return nil
	}

	if flags.ValidateCode {
		if err := validateCodeOnlyResponse(response); err != nil {
			return err
		}
	}

	if flags.NoWrite {
		fmt.Println("\n" + response)
		return nil
	}

	if err := processResponse(response); err != nil {
		fmt.Printf("Input lines: %d\n", totalLength)
		return err
	}

	if planUsed {
		return tickPlanSteps(planPrompt, sentSteps)
	}

	return nil
}

func newClient(cfg *config.Config) api.Client {
	client := &api.ClaudeClient{}
	client.Init(cfg)
	return client
}

func callLLM(requestContent logic.RequestContent, noProgress bool, cfg *config.Config, systemPrompt string) (string, error) {
	client := newClient(cfg)

	if noProgress {
		return client.Call(requestContent, systemPrompt)
	}

	fmt.Printf("Sending request...\n")
	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	response, err := client.Call(requestContent, systemPrompt)

	done <- true
	close(done)

	return response, err
}

func storeBuffer(response string) error {
	if err := logic.PutBuffer(response); err != nil {
		return fmt.Errorf("error writing buffer: %w", err)
	}
	return nil
}

func processResponse(content string) error {
	var parseErrors []string
	elements := logic.ParseCodeFiles(content)

	for _, element := range elements {
		if element.IsCodeFile {
			v := element.CodeFile
			if err := v.Apply(); err != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
			}
		} else {
			if strings.TrimSpace(element.Text) != "" {
				fmt.Println(element.Text)
			}
		}
	}

	if len(parseErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(parseErrors, "; "))
	}

	return nil
}
