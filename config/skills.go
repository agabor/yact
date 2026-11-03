package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SkillName string

const skillDownloadBaseURL = "https://raw.githubusercontent.com/agabor/yact/refs/heads/main/skills/"

const skillsDirName = "skills"

// Names are restricted so that a skill name always maps to a file directly inside the skills folder.
func ParseSkillName(raw string) (SkillName, error) {
	if raw == "" {
		return "", fmt.Errorf("skill name must not be empty")
	}
	if strings.ContainsAny(raw, "/\\") || strings.Contains(raw, "..") {
		return "", fmt.Errorf("invalid skill name %q: must not contain '/', '\\' or '..'", raw)
	}
	return SkillName(raw), nil
}

func getSkillsDir() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, skillsDirName), nil
}

func getSkillFile(name SkillName) (string, error) {
	dir, err := getSkillsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, string(name)+".md"), nil
}

// Skills are attached verbatim, so include directives are not expanded here.
func LoadSkill(name SkillName) (string, error) {
	skillFile, err := getSkillFile(name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(skillFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func DownloadSkill(name SkillName) error {
	fmt.Printf("Downloading skill '%s'...\n", name)

	skillFile, err := getSkillFile(name)
	if err != nil {
		return err
	}

	if _, err := downloadToFile(skillDownloadBaseURL+string(name)+".md", skillFile); err != nil {
		return err
	}

	fmt.Printf("Skill '%s' downloaded successfully.\n", name)

	return nil
}
