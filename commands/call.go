package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"yact/api"
	"yact/config"
	"yact/index"
	"yact/logic"
)

type PlanSelection struct {
	Prompt   string
	Steps    []logic.PlanStep
	AllDone  bool
	Warnings []string
}

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

func indexLevelToVisibility(level int) (string, error) {
	switch level {
	case 1:
		return "public", nil
	case 2:
		return "private", nil
	}
	return "", fmt.Errorf("invalid index level %d, expected 1 or 2", level)
}

func applyIndexOverride(flags config.Flaglist) error {
	if !flags.HasIndex() {
		return logic.RemoveIndexFile()
	}
	visibility, err := indexLevelToVisibility(flags.IndexLevel)
	if err != nil {
		return err
	}
	return index.WriteIndex(visibility)
}

func validateCodeOnlyResponse(content string) error {
	elements, incomplete := logic.ParseCodeFilesDetailed(content)
	if incomplete {
		return fmt.Errorf("response contains an incomplete code block")
	}

	for _, element := range elements {
		switch element.Kind {
		case logic.ElementRejectedFile:
			return fmt.Errorf("response contains a rejected file path: %s", element.Text)
		case logic.ElementText:
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

func printSentFiles(requestContent logic.RequestContent) {
	for _, messagePath := range requestContent.MessagePaths() {
		fmt.Printf("Sending: %s\n", messagePath)
	}
	for _, contextPath := range requestContent.Files {
		fmt.Printf("Sending: %s\n", contextPath)
	}
}

func printSentPrompt(prompt string) {
	if prompt == "" {
		return
	}
	fmt.Println("Sending prompt:")
	fmt.Println(prompt)
}

func printCallResult(result api.CallResult) {
	fmt.Printf("Claude API call took %.2f seconds\n", result.Duration.Seconds())
	fmt.Printf("Files sent: %d\n", result.FileCount)
	fmt.Printf("Token usage - Input: %d, Output: %d\n", result.InputTokens, result.OutputTokens)
	fmt.Printf("Cost: $%.6f\n", result.Cost)

	if result.Truncated {
		fmt.Printf("⚠️  WARNING: Maximum output tokens (%d) reached. Response may be incomplete.\n", config.MaxTokens)
	}
}

func printApplyOutcome(path logic.ProjectPath, outcome logic.ApplyOutcome) {
	switch outcome {
	case logic.OutcomeWritten:
		fmt.Printf("Written: %s\n", path)
	case logic.OutcomeDeleted:
		fmt.Printf("Deleted: %s\n", path)
	case logic.OutcomeNothingToDelete:
		fmt.Printf("Not found (nothing to delete): %s\n", path)
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

func selectPlanSteps(planText string, selector string) (PlanSelection, error) {
	if selector == "" {
		parsedPlan := logic.ParsePlan(planText)
		remainingSteps := parsedPlan.UncheckedSteps()
		if len(parsedPlan.Steps) == len(remainingSteps) {
			return PlanSelection{Prompt: planText, Steps: parsedPlan.Steps}, nil
		}
		if parsedPlan.AllChecked() {
			return PlanSelection{AllDone: true}, nil
		}
		return PlanSelection{
			Prompt: logic.BuildStepsPrompt(parsedPlan, remainingSteps),
			Steps:  remainingSteps,
		}, nil
	}

	stepSelector, err := logic.ParseStepSelector(selector)
	if err != nil {
		return PlanSelection{}, err
	}
	if stepSelector.Kind == logic.StepSelectorWholePlan {
		return PlanSelection{Prompt: planText}, nil
	}

	parsedPlan := logic.ParsePlan(planText)
	selection, err := logic.SelectSteps(parsedPlan, stepSelector)
	if err != nil {
		return PlanSelection{}, err
	}

	if selection.AllDone {
		return PlanSelection{AllDone: true, Warnings: selection.Warnings}, nil
	}

	return PlanSelection{
		Prompt:   logic.BuildStepsPrompt(parsedPlan, selection.Steps),
		Steps:    selection.Steps,
		Warnings: selection.Warnings,
	}, nil
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

func submitBatch(requestContent logic.RequestContent, cfg *config.Config, systemPrompt string, mode config.OutputMode) error {
	client := newClient(cfg)

	batchID, err := client.SubmitBatch(requestContent, systemPrompt)
	if err != nil {
		return err
	}

	job := logic.BatchJob{
		ID:   batchID,
		Mode: mode,
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
		projectPath, err := logic.ParseProjectPath(contextPath)
		if err != nil {
			return 0, err
		}
		codeFile, err := logic.ReadAsFile(projectPath)
		if err != nil {
			return 0, err
		}
		totalLength += len(strings.Split(codeFile.Content, "\n"))
	}

	return totalLength, nil
}

func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, prompt string) error {
	mode, err := config.ParseOutputMode(flags)
	if err != nil {
		return err
	}

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
		selection, err := selectPlanSteps(planPrompt, flags.Steps)

		if err != nil {
			return err
		}
		printWarnings(selection.Warnings)
		if selection.AllDone {
			fmt.Println("All plan steps are done.")
			return nil
		}
		if len(selection.Steps) > 0 {
			fmt.Printf("Plan steps: %s\n", logic.FormatNumbers(logic.StepNumbers(selection.Steps)))
		}
		sentSteps = selection.Steps
		requestContentForLLM.Prompt = selection.Prompt
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

	if flags.Verbose {
		printSentFiles(requestContentForLLM)
		printSentPrompt(requestContentForLLM.Prompt)
	}

	if flags.Dry {
		fmt.Println("Dry run: LLM call skipped.")
		return nil
	}

	if flags.Batch {
		return submitBatch(requestContentForLLM, cfg, systemPrompt, mode)
	}

	result, err := callLLM(requestContentForLLM, flags.NoProgress, cfg, systemPrompt)
	if err != nil {
		return err
	}

	printCallResult(result)

	if err := storeBuffer(result.Text); err != nil {
		return err
	}

	if err := deliverResponse(mode, result.Text); err != nil {
		fmt.Printf("Input lines: %d\n", totalLength)
		return err
	}

	// A step whose output was only printed or written as a plan stays unchecked.
	if planUsed && (mode == config.OutputApply || mode == config.OutputValidateAndApply) {
		return tickPlanSteps(planPrompt, sentSteps)
	}

	return nil
}

func newClient(cfg *config.Config) api.Client {
	client := &api.ClaudeClient{}
	client.Init(cfg)
	return client
}

func callLLM(requestContent logic.RequestContent, noProgress bool, cfg *config.Config, systemPrompt string) (api.CallResult, error) {
	client := newClient(cfg)

	if noProgress {
		return client.Call(requestContent, systemPrompt)
	}

	fmt.Printf("Sending request...\n")
	fmt.Printf("Model: %s\n", client.GetModelName())

	done := make(chan bool)
	go showProgress(done)

	result, err := client.Call(requestContent, systemPrompt)

	done <- true
	close(done)

	return result, err
}

func storeBuffer(response string) error {
	if err := logic.PutBuffer(response); err != nil {
		return fmt.Errorf("error writing buffer: %w", err)
	}
	return nil
}

func deliverResponse(mode config.OutputMode, response string) error {
	switch mode {
	case config.OutputWritePlan:
		if err := logic.WritePlan(response); err != nil {
			return err
		}
		fmt.Printf("Plan written to %s\n", config.GetProjectPromptPath())
		return nil
	case config.OutputPrintOnly:
		fmt.Println("\n" + response)
		return nil
	case config.OutputValidateAndApply:
		if err := validateCodeOnlyResponse(response); err != nil {
			return err
		}
		return processResponse(response)
	case config.OutputApply:
		return processResponse(response)
	}
	return fmt.Errorf("unknown output mode %d", mode)
}

func processResponse(content string) error {
	var parseErrors []string
	elements, incomplete := logic.ParseCodeFilesDetailed(content)
	if incomplete {
		fmt.Println("Warning: Incomplete code block.")
	}

	for _, element := range elements {
		switch element.Kind {
		case logic.ElementCodeFile:
			v := element.CodeFile
			outcome, err := v.Apply()
			if err != nil {
				parseErrors = append(parseErrors, fmt.Sprintf("%v", err))
				continue
			}
			printApplyOutcome(v.Path, outcome)
		case logic.ElementRejectedFile:
			parseErrors = append(parseErrors, element.Text)
		case logic.ElementText:
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
