package commands

import (
	"fmt"
	"strings"

	"yact/logic"
)

func isReferencedInPrompt(prompt string, path string) bool {
	return strings.Contains(prompt, path)
}

func splitContextByReference(context []string, prompt string) ([]string, []string) {
	kept := make([]string, 0, len(context))
	removed := make([]string, 0)

	for _, path := range context {
		if isReferencedInPrompt(prompt, path) {
			kept = append(kept, path)
		} else {
			removed = append(removed, path)
		}
	}

	return kept, removed
}

func HandleNarrowCommand() error {
	transaction, err := logic.LoadTransaction()
	if err != nil {
		return err
	}

	prompt := strings.Join(transaction.Request, "\n")

	kept, removed := splitContextByReference(transaction.Context, prompt)

	for _, path := range removed {
		fmt.Printf("Removed: %s\n", path)
	}

	if len(removed) == 0 {
		fmt.Println("Nothing to remove")
		return nil
	}

	transaction.Context = kept

	return transaction.Save()
}