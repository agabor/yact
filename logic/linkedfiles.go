package logic

import (
	"slices"

	"yact/config"
	"yact/projectpath"
)

// LinkedFiles are the only paths one response may write, so the LLM cannot change files the prompt did not name.
type LinkedFiles []projectpath.Path

// feedback.md carries a message to the user rather than project content, so it needs no link.
func (l LinkedFiles) Allows(path projectpath.Path) bool {
	if path == FeedbackFilePath() {
		return true
	}
	return slices.Contains(l, path)
}

func FeedbackFilePath() projectpath.Path {
	return projectpath.Join(config.GetProjectTempDir(), "feedback.md")
}
