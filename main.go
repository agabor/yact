package main

import (
	"fmt"
	"os"
	"strconv"
	"yact/config/systemprompt"

	"yact/commands"
	"yact/config"

	flag "github.com/spf13/pflag"
)

func requireArgCount(command string, args []string, validCounts ...int) {
	for _, count := range validCounts {
		if len(args) == count {
			return
		}
	}
	if len(validCounts) == 1 {
		fmt.Fprintf(os.Stderr, "Error: %s command requires %d argument(s)\n", command, validCounts[0])
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s command received %d argument(s), expected one of: %v\n", command, len(args), validCounts)
	}
	os.Exit(1)
}

func requireNoArgs(command string, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "Error: %s command takes no arguments\n", command)
		os.Exit(1)
	}
}

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
		requireArgCount("config", commandArgs, 0, 2)
		commandErr = commands.HandleConfigCommand(commandArgs, cfg)
	case "act":
		requireNoArgs("act", commandArgs)
		commandErr = commands.HandleActCommand(*thinkFlag, cfg, systemprompt.Act)
	case "snip":
		requireArgCount("snip", commandArgs, 4)
		inputFile := commandArgs[0]
		startLine, err := strconv.Atoi(commandArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid start line number: %v\n", err)
			os.Exit(1)
		}
		endLine, err := strconv.Atoi(commandArgs[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid end line number: %v\n", err)
			os.Exit(1)
		}
		prompt := commandArgs[3]
		commandErr = commands.HandleSnipCommand(inputFile, startLine, endLine, prompt, *thinkFlag, cfg, systemprompt.Snip)
	case "ask":
		requireNoArgs("ask", commandArgs)
		commandErr = commands.HandleAskCommand(*thinkFlag, cfg, systemprompt.Ask)
	case "plan":
		requireNoArgs("plan", commandArgs)
		commandErr = commands.HandlePlanCommand(*thinkFlag, cfg, systemprompt.Plan)
	case "context":
		requireNoArgs("context", commandArgs)
		commandErr = commands.HandleContextCommand(cfg)
	case "index":
		requireNoArgs("index", commandArgs)
		commandErr = commands.HandleIndexCommand(cfg)
	case "new":
		commandErr = commands.HandleNewCommand()
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