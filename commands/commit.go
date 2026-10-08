package commands

import (
	"fmt"
	"strings"

	"yact/logic"
)

func HandleCommitCommand() error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements, incomplete := logic.ParseCodeFilesDetailed(content)
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