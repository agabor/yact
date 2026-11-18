package commands

import (
	"errors"
	"fmt"
	"strings"

	"yact/config"
	"yact/logic"
	"yact/projectpath"
)

func HandlePromptCommand(text string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("prompt text is empty")
	}

	links, err := logic.ResolvePromptLinks(text)
	if err != nil {
		return err
	}
	printWarnings(unresolvedFileWarnings(links.UnresolvedPaths))

	keywordPaths, err := resolvePromptKeywordFiles(text)
	if err != nil {
		return err
	}

	paths := append([]projectpath.Path{}, links.Paths...)
	paths = append(paths, keywordPaths...)

	existingPrompt, _, err := logic.ReadPlan()
	if err != nil {
		return err
	}

	updatedPrompt := logic.AppendPromptBody(
		logic.AddReadLinksToPrompt(existingPrompt, logic.UniquePaths(paths)),
		logic.StripKeywordMarkers(text),
	)

	if err := logic.WritePlan(updatedPrompt); err != nil {
		return err
	}

	fmt.Printf("Prompt appended to %s\n", config.GetProjectPromptPath())
	return nil
}
