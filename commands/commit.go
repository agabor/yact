package commands

import (
	"fmt"
	"strings"

	"yact/logic"
)

func printFreeText(elements []logic.ParsedElement) {
	for _, element := range elements {
		if element.Kind == logic.ElementText && strings.TrimSpace(element.Text) != "" {
			fmt.Println(element.Text)
		}
	}
}

func parseBufferElements(content string, validateCode bool) ([]logic.ParsedElement, bool, error) {
	if validateCode {
		if err := validateCodeOnlyResponse(content); err != nil {
			return nil, false, err
		}
	}

	elements, incomplete := logic.ParseCodeFilesDetailed(content)
	printFreeText(elements)
	return elements, incomplete, nil
}

func HandleCommitCommand(validateCode bool) error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements, incomplete, err := parseBufferElements(content, validateCode)
	if err != nil {
		return err
	}
	if incomplete {
		fmt.Println("Warning: Incomplete code block.")
	}

	var writeErrors []string

	for _, element := range elements {
		switch element.Kind {
		case logic.ElementCodeFile:
			codeFile := element.CodeFile
			outcome, err := codeFile.Apply()
			if err != nil {
				writeErrors = append(writeErrors, fmt.Sprintf("%v", err))
				continue
			}
			printApplyOutcome(codeFile.Path, outcome)
		case logic.ElementRejectedFile:
			writeErrors = append(writeErrors, element.Text)
		}
	}

	if len(writeErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(writeErrors, "; "))
	}

	return nil
}