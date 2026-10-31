package commands

import (
	"fmt"
	"os"
	"strings"

	"yact/logic"
)

// Files longer than this are easy to overlook in a request, so they are flagged before the call.
const longFileLineThreshold = 300

type fileLineCount struct {
	Path  string
	Lines int
}

func countInputLines(requestContent logic.RequestContent) (int, []fileLineCount, error) {
	totalLength := 0
	var fileCounts []fileLineCount

	if requestContent.Prompt != "" {
		totalLength += len(strings.Split(requestContent.Prompt, "\n"))
	}

	for _, messagePath := range requestContent.MessagePaths() {
		messageContent, err := os.ReadFile(messagePath)
		if err != nil {
			return 0, nil, err
		}
		lines := len(strings.Split(string(messageContent), "\n"))
		totalLength += lines
		fileCounts = append(fileCounts, fileLineCount{Path: messagePath, Lines: lines})
	}

	for _, contextPath := range requestContent.Files {
		projectPath, err := logic.ParseProjectPath(contextPath)
		if err != nil {
			return 0, nil, err
		}
		codeFile, err := logic.ReadAsFile(projectPath)
		if err != nil {
			return 0, nil, err
		}
		lines := len(strings.Split(codeFile.Content, "\n"))
		totalLength += lines
		fileCounts = append(fileCounts, fileLineCount{Path: contextPath, Lines: lines})
	}

	return totalLength, fileCounts, nil
}

func longFileWarnings(counts []fileLineCount) []string {
	var warnings []string
	for _, count := range counts {
		if count.Lines > longFileLineThreshold {
			warnings = append(warnings, fmt.Sprintf("%s has %d lines, more than %d", count.Path, count.Lines, longFileLineThreshold))
		}
	}
	return warnings
}