package main

import (
	"fmt"
	"os"
	"yact/config/systemprompt"

	"yact/commands"
	"yact/config"

	flag "github.com/spf13/pflag"
)

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	thinkFlag := flag.BoolP("think", "t", false, "Enable Claude's extended thinking mode")

	flag.Parse()

	args := flag.Args()

	if *helpFlag {
		commands.ShowHelp()
		return
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no command provided\n")
		fmt.Println("Run 'y --help' for usage information.")
		os.Exit(1)
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
		if len(commandArgs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: act command requires a prompt\n")
			os.Exit(1)
		}
		commandErr = commands.HandleActCommand(commandArgs[0], *thinkFlag, cfg, systemprompt.Act)
	case "bash":
		if len(commandArgs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: bash command requires a prompt\n")
			os.Exit(1)
		}
		commandErr = commands.HandleActCommand(commandArgs[0], *thinkFlag, cfg, systemprompt.Bash)
	case "ask":
		if len(commandArgs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: ask command requires a prompt\n")
			os.Exit(1)
		}
		commandErr = commands.HandleAskCommand(commandArgs[0], *thinkFlag, cfg, systemprompt.Ask)
	case "plan":
		if len(commandArgs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: plan command requires a prompt\n")
			os.Exit(1)
		}
		commandErr = commands.HandlePlanCommand(commandArgs[0], *thinkFlag, cfg, systemprompt.Plan)
	case "new":
		commandErr = commands.HandleNewCommand()
	case "step":
		if len(commandArgs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: step index required\n")
			os.Exit(1)
		}
		prompt := "implement step " + commandArgs[0] + ". Make no other changes."
		commandErr = commands.HandleActCommand(prompt, *thinkFlag, cfg, systemprompt.Act)
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
