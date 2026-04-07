package commands

import (
	"fmt"
	"strings"
	"yact/logic"
)

func HandleRestoreCommand() error {
	transactions, err := logic.LoadContext()
	if err != nil {
		return err
	}

	fileVersions := make(map[string]string)
	var restoreErrors []string

	for _, transaction := range transactions {
		for _, file := range transaction.Context {
			fileVersions[file.Path()] = file.Content
		}

		if transaction.Type == logic.TransactionTypeAct {
			blocks, _ := logic.ParseCodeBlocks(transaction.Response)
			for _, block := range blocks {
				fileVersions[block.Path()] = block.Content
			}
		}
	}

	for path, content := range fileVersions {
		codeFile := logic.NewCodeFile(path, content)
		err := codeFile.Write()
		if err != nil {
			restoreErrors = append(restoreErrors, fmt.Sprintf("%v", err))
		}
	}

	if len(restoreErrors) > 0 {
		return fmt.Errorf("error restoring files: %s", strings.Join(restoreErrors, "; "))
	}

	fmt.Println("Files restored to last known state")
	return nil
}
