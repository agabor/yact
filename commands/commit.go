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

func parseBufferElements(content string) []logic.ParsedElement {
	elements := logic.ParseCodeFilesDetailed(content)
	printFreeText(elements)
	return elements
}

func HandleCommitCommand() error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements := parseBufferElements(content)

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
		case logic.ElementText:
			fmt.Println(element.Text)
		}
	}

	if len(writeErrors) > 0 {
		return fmt.Errorf("error processing code blocks: %s", strings.Join(writeErrors, "; "))
	}

	return nil
}
