package commands

import (
	"fmt"

	"yact/logic"
)

func HandleBufferCommand() error {
	content, err := logic.ReadBuffer()
	if err != nil {
		return err
	}
	fmt.Println(content)
	return nil
}
