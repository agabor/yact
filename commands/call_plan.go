package commands

import (
	"fmt"
	"strings"
	"yact/logic"
	"yact/projectpath"
)

const longFileLineThreshold = 300

// The read link block above the separator is never sent as text, so only the body is sent, with keyword markers removed.
func sentPromptText(promptText string) string {
	return logic.StripKeywordMarkers(logic.SplitPromptFooter(promptText).Body)
}

func applyPromptLinks(requestContent *logic.RequestContent) error {
	if requestContent.Prompt == "" {
		return nil
	}

	links, err := logic.ResolvePromptLinks(requestContent.Prompt)
	if err != nil {
		return err
	}

	for _, path := range links.UnresolvedPaths {
		fmt.Printf("Unresolved: %s\n", path)
	}

	keywordPaths, err := resolvePromptKeywordFiles(requestContent.Prompt)
	if err != nil {
		return err
	}

	paths := append([]projectpath.Path{}, links.Paths...)
	paths = append(paths, keywordPaths...)
	for _, path := range logic.UniquePaths(paths) {
		file, err := logic.ReadAsFile(path)
		if err != nil {
			return err
		}
		lineCount := strings.Count(file.Content, "\n") + 1
		if lineCount > longFileLineThreshold {
			fmt.Printf("Reading: %s (%d lines) !!!\n", path, lineCount)
		} else {
			fmt.Printf("Reading: %s (%d lines)\n", path, lineCount)
		}
		requestContent.Files = append(requestContent.Files, file)
	}
	requestContent.Prompt = sentPromptText(requestContent.Prompt)
	return nil
}
