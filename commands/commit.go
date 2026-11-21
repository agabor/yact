package commands

import (
	"yact/logic"
)

func HandleCommitCommand() error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements, err := logic.ParseCodeFilesDetailed(content)
	if err != nil {
		return err
	}
	return processResponse(elements)
}
