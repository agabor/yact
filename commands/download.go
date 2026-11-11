package commands

import (
	"errors"
	"fmt"

	"yact/config"
)

const downloadUsage = "usage: y download prompt <command> | skill <name>"

// DownloadTarget is either a system prompt or a skill, so a download always names one of them.
type DownloadTarget interface {
	isDownloadTarget()
}

type PromptDownload struct {
	Name config.PromptName
}

func (PromptDownload) isDownloadTarget() {}

type SkillDownload struct {
	Name config.SkillName
}

func (SkillDownload) isDownloadTarget() {}

func parseDownloadArgs(args []string) (DownloadTarget, error) {
	if len(args) != 2 {
		return nil, errors.New(downloadUsage)
	}

	switch args[0] {
	case "prompt":
		name, err := config.ParsePromptName(args[1])
		if err != nil {
			return nil, err
		}
		return PromptDownload{Name: name}, nil
	case "skill":
		name, err := config.ParseSkillName(args[1])
		if err != nil {
			return nil, err
		}
		return SkillDownload{Name: name}, nil
	}
	return nil, errors.New(downloadUsage)
}

// This is the only place that performs an explicit download, so that both kinds of target
// replace the local copy the same way.
func downloadTarget(target DownloadTarget) error {
	switch selected := target.(type) {
	case PromptDownload:
		return config.DownloadPrompt(selected.Name)
	case SkillDownload:
		return config.DownloadSkill(selected.Name)
	}
	return fmt.Errorf("unknown download target %T", target)
}

func HandleDownloadCommand(args []string) error {
	target, err := parseDownloadArgs(args)
	if err != nil {
		return err
	}
	return downloadTarget(target)
}
