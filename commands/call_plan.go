package commands

import (
	"strings"

	"yact/logic"
	"yact/projectpath"
)

func combinePromptText(promptFileText string, cliPrompt string) string {
	sections := logic.SplitPromptFooter(promptFileText)
	sections.Body = joinPromptBody(sections.Body, cliPrompt)
	return logic.JoinPromptSections(sections)
}

func joinPromptBody(fileBody string, cliPrompt string) string {
	fileText := strings.TrimSpace(fileBody)
	cliText := strings.TrimSpace(cliPrompt)
	if fileText == "" {
		return cliText
	}
	if cliText == "" {
		return fileText
	}
	return fileText + "\n\n" + cliText
}

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

	if len(links.UnresolvedPaths) > 0 {
		printWarnings(unresolvedFileWarnings(links.UnresolvedPaths))
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
