package main

import (
	"fmt"
	"io"
	"os"
	"yact/config/systemprompt"
	"yact/logic"

	"yact/commands"
	"yact/config"

	flag "github.com/spf13/pflag"
)

func isStdinPiped() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func getPromptFromStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	safeFlag := flag.BoolP("safe", "s", false, "Add .new suffix to generated files")
	thinkFlag := flag.BoolP("think", "t", false, "Enable Claude's extended thinking mode")

	flag.Parse()

	args := flag.Args()

	if *helpFlag {
		commands.ShowHelp()
		return
	}

	if len(args) == 0 {
		if isStdinPiped() {
			stdinContent, err := getPromptFromStdin()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(stdinContent)
			args = []string{"act", stdinContent}
		} else {
			fmt.Fprintf(os.Stderr, "Error: no command provided\n")
			fmt.Println("Run 'y --help' for usage information.")
			os.Exit(1)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	command := args[0]
	commandArgs := []string{}
	if len(args) > 1 {
		commandArgs = args[1:]
	}

	var commandErr error

	switch command {
	case "help":
		commands.ShowHelp()
		return
	case "read":
		commandErr = commands.HandleReadCommand(commandArgs)
	case "config":
		commandErr = commands.HandleConfigCommand(commandArgs, cfg)
	case "context":
		if len(commandArgs) != 0 {
			fmt.Fprintf(os.Stderr, "Error: the go command takes no arguments\n")
			os.Exit(1)
		}
		commandErr = commands.HandleContextCommand()
	case "pop":
		commandErr = commands.HandlePop(commandArgs)
	case "reset":
		if len(commandArgs) != 0 {
			fmt.Fprintf(os.Stderr, "Error: the reset command takes no arguments\n")
			os.Exit(1)
		}
		commandErr = commands.HandleResetCommand()
	case "restore":
		if len(commandArgs) != 0 {
			fmt.Fprintf(os.Stderr, "Error: the restore command takes no arguments\n")
			os.Exit(1)
		}
		commandErr = commands.HandleRestoreCommand()
	case "act":
		commandErr = commands.HandleActCommand(commandArgs, *safeFlag, *thinkFlag, cfg, systemprompt.Act)
	case "bash":
		commandErr = commands.HandleActCommand(commandArgs, *safeFlag, *thinkFlag, cfg, systemprompt.Bash)
	case "ask":
		commandErr = commands.HandleVerbalCommand(commandArgs, *thinkFlag, cfg, systemprompt.Ask, logic.TransactionTypeQuestion)
	case "plan":
		commandErr = commands.HandleVerbalCommand(commandArgs, *thinkFlag, cfg, systemprompt.Plan, logic.TransactionTypePlan)
	case "new":
		commandErr = commands.HandleNewCommand()
	case "step":
		if len(commandArgs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: step index required\n")
			os.Exit(1)
		}
		stepArgs := append([]string{"implement", "step"}, commandArgs...)
		stepArgs = append(stepArgs, ". Make no other changes.")
		commandErr = commands.HandleActCommand(stepArgs, *safeFlag, *thinkFlag, cfg, systemprompt.Act)
	case "go":
		if len(commandArgs) != 0 {
			fmt.Fprintf(os.Stderr, "Error: the go command takes no arguments\n")
			os.Exit(1)
		}
		commandErr = commands.HandleGoCommand(*thinkFlag, cfg, systemprompt.Act)
	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		fmt.Println("Run 'y --help' for usage information.")
		os.Exit(1)
	}

	if commandErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		os.Exit(1)
	}
}
