package commands

import (
	"yact/logic"
)

func HandleCommitCommand() error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements := logic.ParseCodeFilesDetailed(content)
	return processResponse(elements)
}
