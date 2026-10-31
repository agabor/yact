package commands

import (
	"fmt"
	"strings"

	"yact/config"
	"yact/logic"
)

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