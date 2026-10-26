package commands

import (
	"fmt"
	"strconv"

	"yact/config"
)

const indexFilePath = ".yact/index.txt"

func parseIndexLevel(arg string) (int, error) {
	level, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid index level %q, expected 0, 1 or 2", arg)
	}
	switch level {
	case 0, 1, 2:
		return level, nil
	}
	return 0, fmt.Errorf("invalid index level %d, expected 0, 1 or 2", level)
}

func HandleIndexCommand(args []string) error {
	level, err := parseIndexLevel(args[0])
	if err != nil {
		return err
	}

	flags := config.Flaglist{IndexLevel: level}
	if err := applyIndexOverride(flags); err != nil {
		return err
	}

	if level == 0 {
		fmt.Printf("Index file removed: %s\n", indexFilePath)
		return nil
	}
	fmt.Printf("Index written to %s\n", indexFilePath)
	return nil
}
