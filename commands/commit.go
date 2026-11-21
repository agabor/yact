package commands

import (
	"yact/logic"
)

func HandleCommitCommand() error {
	session, err := logic.ReadBuffer()
	if err != nil {
		return err
	}

	elements, err := logic.ParseCodeFilesDetailed(session.Response, session.Linked)
	if err != nil {
		return err
	}
	return processResponse(elements)
}
