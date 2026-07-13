// CLI entry point that parses commands, manages flags, and delegates to command handlers
// CLI entry point that parses commands, manages flags, and delegates to command handlers
package main

import (
	"fmt"
	"os"
	"strconv"

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

func getModelOverride(fable, opus, sonnet, haiku bool) string {
	if fable {
		return "fable"
	}
	if opus {
		return "opus"
	}
	if sonnet {
		return "sonnet"
	}
	if haiku {
		return "haiku"
	}
	return ""
}

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	thinkFlag := flag.BoolP("think", "t", false, "Enable Claude's extended thinking mode")
	noWriteFlag := flag.BoolP("no-write", "n", false, "Do not write files, print response instead")
	fableFlag := flag.BoolP("fable", "f", false, "Use Claude Fable model")
	opusFlag := flag.BoolP("opus", "o", false, "Use Claude Opus model")
	sonnetFlag := flag.BoolP("sonnet", "s", false, "Use Claude Sonnet model")
	haikuFlag := flag.Bool("haiku", false, "Use Claude Haiku model")

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

	modelOverride := getModelOverride(*fableFlag, *opusFlag, *sonnetFlag, *haikuFlag)

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
	case "prompt":
		requireArgCount("prompt", commandArgs, 1)
		commandErr = commands.HandlePromptCommand(commandArgs)
	case "act":
		requireNoArgs("act", commandArgs)
		actPrompt, err := config.LoadPrompt("act")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading act prompt: %v\n", err)
			os.Exit(1)
		}
		commandErr = commands.HandleActCommand(*thinkFlag, *noWriteFlag, cfg, actPrompt, modelOverride)
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
		snippetPrompt, err := config.LoadPrompt("snip")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading snip prompt: %v\n", err)
			os.Exit(1)
		}
		commandErr = commands.HandleSnipCommand(inputFile, startLine, endLine, prompt, *thinkFlag, cfg, snippetPrompt)
	case "ask":
		requireNoArgs("ask", commandArgs)
		askPrompt, err := config.LoadPrompt("ask")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading ask prompt: %v\n", err)
			os.Exit(1)
		}
		commandErr = commands.HandleAskCommand(*thinkFlag, cfg, askPrompt, modelOverride)
	case "bash":
		requireNoArgs("bash", commandArgs)
		bashPrompt, err := config.LoadPrompt("bash")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading bash prompt: %v\n", err)
			os.Exit(1)
		}
		commandErr = commands.HandleBashCommand(*thinkFlag, cfg, bashPrompt, modelOverride)
	case "plan":
		requireNoArgs("plan", commandArgs)
		planPrompt, err := config.LoadPrompt("plan")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading plan prompt: %v\n", err)
			os.Exit(1)
		}
		commandErr = commands.HandlePlanCommand(*thinkFlag, cfg, planPrompt, modelOverride)
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