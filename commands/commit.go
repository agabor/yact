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

func parseBufferElements(content string, validateCode bool) ([]logic.ParsedElement, error) {
	if validateCode {
		if err := validateCodeOnlyResponse(content); err != nil {
			return nil, err
		}
	}

	elements := logic.ParseCodeFiles(content)
	printFreeText(elements)
	return elements, nil
}

func HandleCommitCommand(validateCode bool) error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements, err := parseBufferElements(content, validateCode)
	if err != nil {
		return err
	}

	var writeErrors []string

	for _, element := range elements {
		switch element.Kind {
		case logic.ElementCodeFile:
			codeFile := element.CodeFile
			if err := codeFile.Apply(); err != nil {
				writeErrors = append(writeErrors, fmt.Sprintf("%v", err))
			}
		case logic.ElementRejectedFile:
			writeErrors = append(writeErrors, element.Text)
		}
	}

	if len(writeErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(writeErrors, "; "))
	}

	return nil
}