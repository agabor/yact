package commands

import (
	"fmt"
	"strings"

	"yact/logic"
)

func validateBufferCodeOnly(elements []logic.ParsedElement, incomplete bool) error {
	if incomplete {
		return fmt.Errorf("buffer contains an incomplete code block")
	}

	for _, element := range elements {
		if !element.IsCodeFile {
			if strings.TrimSpace(element.Text) != "" {
				return fmt.Errorf("buffer contains free text outside code blocks")
			}
		}
	}

	return nil
}

func printFreeText(elements []logic.ParsedElement) {
	for _, element := range elements {
		if !element.IsCodeFile && strings.TrimSpace(element.Text) != "" {
			fmt.Println(element.Text)
		}
	}
}

func parseBufferElements(content string, validateCode bool) ([]logic.ParsedElement, error) {
	if validateCode {
		elements, incomplete := logic.ParseCodeFilesDetailed(content)
		if err := validateBufferCodeOnly(elements, incomplete); err != nil {
			return nil, err
		}
		return elements, nil
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
		if element.IsCodeFile {
			codeFile := element.CodeFile
			if codeFile.IsEmpty() {
				if err := codeFile.Delete(); err != nil {
					writeErrors = append(writeErrors, fmt.Sprintf("%v", err))
				}
			} else {
				if err := codeFile.Write(); err != nil {
					writeErrors = append(writeErrors, fmt.Sprintf("%v", err))
				}
			}
		}
	}

	if len(writeErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(writeErrors, "; "))
	}

	return nil
}