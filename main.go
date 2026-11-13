package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"yact/commands"
	"yact/config"

	flag "github.com/spf13/pflag"
)

// Set at build time with -ldflags "-X main.version=<tag>"; local builds keep "dev".
var version = "dev"

func requireArgCount(command string, args []string, validCounts ...int) {
	if slices.Contains(validCounts, len(args)) {
		return
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

func getOptionalPrompt(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

func promptNotFoundMessage(command config.PromptName, err error) {
	fmt.Fprintf(os.Stderr, "Error: no system prompt found for command '%s'\n", command)
	fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
	fmt.Fprintln(os.Stderr, "Run 'y --help' for usage information.")
}

func downloadAndLoadSystemPrompt(command config.PromptName) (string, config.PromptDefaults, error) {
	if err := config.DownloadPrompt(command); err != nil {
		return "", config.PromptDefaults{}, err
	}
	return config.LoadPromptWithFlags(command)
}

func loadSystemPrompt(command config.PromptName) (string, config.PromptDefaults, error) {
	systemPrompt, defaults, err := config.LoadPromptWithFlags(command)
	if errors.Is(err, fs.ErrNotExist) {
		return downloadAndLoadSystemPrompt(command)
	}
	return systemPrompt, defaults, err
}

func main() {
	cli := config.RegisterFlags(flag.CommandLine)
	flag.Parse()

	flags, err := cli.ToFlaglist(flag.CommandLine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if cli.Help {
		commands.ShowHelp()
		return
	}

	if cli.Version {
		fmt.Printf("yact %s\n", version)
		return
	}

	args := flag.Args()

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
	case "output":
		requireNoArgs("output", commandArgs)
		commandErr = commands.HandleOutputCommand()
	case "keyword":
		requireArgCount("keyword", commandArgs, 1)
		commandErr = commands.HandleKeywordCommand(commandArgs)
	case "config":
		requireArgCount("config", commandArgs, 0, 2)
		commandErr = commands.HandleConfigCommand(commandArgs, cfg)
	case "new":
		commandErr = commands.HandleNewCommand()
	case "commit":
		requireNoArgs("commit", commandArgs)
		commandErr = commands.HandleCommitCommand(cli.Strict)
	case "query":
		requireArgCount("query", commandArgs, 0, 1)
		resolved, err := config.ResolveFlags(flags, config.QueryDefaults())
		if err != nil {
			commandErr = err
			break
		}
		if flagString := resolved.FlagString(); flagString != "" {
			fmt.Printf("Executing: y %s query\n", flagString)
		}
		commandErr = commands.HandleCommand(resolved, cfg, "", getOptionalPrompt(commandArgs))
	case "batch":
		requireNoArgs("batch", commandArgs)
		commandErr = commands.HandleBatchCommand(cfg)
	case "download":
		commandErr = commands.HandleDownloadCommand(commandArgs)
	case "index":
		requireArgCount("index", commandArgs, 1)
		commandErr = commands.HandleIndexCommand(commandArgs)
	case "compact":
		requireNoArgs("compact", commandArgs)
		commandErr = commands.HandleCompactCommand()
	default:
		requireArgCount(command, commandArgs, 0, 1)

		promptName, err := config.ParsePromptName(command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		systemPrompt, defaults, err := loadSystemPrompt(promptName)
		if err != nil {
			promptNotFoundMessage(promptName, err)
			os.Exit(1)
		}

		resolved, err := config.ResolveFlags(flags, defaults)
		if err != nil {
			commandErr = err
			break
		}

		cleanedPrompt := systemPrompt
		if strings.HasPrefix(cleanedPrompt, "flags: ") {
			lines := strings.Split(cleanedPrompt, "\n")
			if len(lines) > 1 {
				cleanedPrompt = strings.Join(lines[1:], "\n")
			}
		}

		if flagString := resolved.FlagString(); flagString != "" {
			fmt.Printf("Executing: y %s %s\n", flagString, command)
		}
		commandErr = commands.HandleCommand(resolved, cfg, cleanedPrompt, getOptionalPrompt(commandArgs))
	}

	if commandErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		os.Exit(1)
	}
}
