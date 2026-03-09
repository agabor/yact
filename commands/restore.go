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
			path := strings.TrimPrefix(strings.TrimSpace(file.Path), "./")
			fileVersions[path] = file.Content
		}

		if transaction.Type == logic.TransactionTypeAct {
			blocks, _, _ := logic.ParseCodeBlocks(transaction.Response)
			for _, block := range blocks {
				path := strings.TrimPrefix(strings.TrimSpace(block.Path), "./")
				fileVersions[path] = block.Content
			}
		}
	}

	for path, content := range fileVersions {
		codeFile := logic.CodeFile{Path: path, Content: content}
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