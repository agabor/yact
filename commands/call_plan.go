package commands

import (
	"fmt"
	"yact/logic"
	"yact/projectpath"
)

const longFileLineThreshold = 300

// The read link block above the separator is never sent as text, so only the body is sent, with keyword markers removed.
func sentPromptText(promptText string) string {
	return logic.StripKeywordMarkers(logic.SplitPromptFooter(promptText).Body)
}

func applyPromptLinks(requestContent logic.RequestContent) (logic.RequestContent, error) {
	if requestContent.Prompt == "" {
		return requestContent, nil
	}

	links, err := logic.ResolvePromptLinks(requestContent.Prompt)
	if err != nil {
		return requestContent, err
	}

	for _, path := range links.Paths {
		fmt.Printf("Reading: %s\n", path)
	}

	for _, path := range links.UnresolvedPaths {
		fmt.Printf("Unresolved: %s\n", path)
	}

	keywordPaths, err := resolvePromptKeywordFiles(requestContent.Prompt)
	if err != nil {
		return requestContent, err
	}

	paths := append([]projectpath.Path{}, links.Paths...)
	paths = append(paths, keywordPaths...)
	requestContent.Files = logic.UniquePaths(paths)
	requestContent.Prompt = sentPromptText(requestContent.Prompt)
	return requestContent, nil
}
