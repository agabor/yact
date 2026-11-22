package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"yact/api"
	"yact/commands"
	"yact/config"

	flag "github.com/spf13/pflag"
)

// Set at build time with -ldflags "-X main.version=<tag>"; local builds keep "dev".
var version = "dev"

func requireArgCount(command string, args []string, validCounts ...int) error {
	if slices.Contains(validCounts, len(args)) {
		return nil
	}
	if len(validCounts) == 1 {
		return fmt.Errorf("%s command requires %d argument(s)", command, validCounts[0])
	}
	return fmt.Errorf("%s command received %d argument(s), expected one of: %v", command, len(args), validCounts)
}

func requireNoArgs(command string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("%s command takes no arguments", command)
	}
	return nil
}

func promptNotFoundError(command config.PromptName, err error) error {
	return fmt.Errorf("no system prompt found for command '%s'\nDownload failed: %v\nRun 'y --help' for usage information.", command, err)
}

func downloadAndLoadSystemPrompt(command config.PromptName) (config.SystemPrompt, error) {
	if err := config.DownloadPrompt(command); err != nil {
		return config.SystemPrompt{}, err
	}
	return config.LoadPrompt(command)
}

func loadSystemPrompt(command config.PromptName) (config.SystemPrompt, error) {
	systemPrompt, err := config.LoadPrompt(command)
	if errors.Is(err, fs.ErrNotExist) {
		return downloadAndLoadSystemPrompt(command)
	}
	return systemPrompt, err
}

func printExecuting(resolved config.Flaglist, command string) {
	if flagString := resolved.FlagString(); flagString != "" {
		fmt.Printf("Executing: y %s %s\n", flagString, command)
	}
}

func runPromptCommand(command string, flags config.Flaglist, cfg *config.Config, newClient api.ClientFactory) error {
	promptName, err := config.ParsePromptName(command)
	if err != nil {
		return err
	}

	systemPrompt, err := loadSystemPrompt(promptName)
	if err != nil {
		return promptNotFoundError(promptName, err)
	}

	resolved := config.ResolveFlags(flags, systemPrompt.Defaults)
	printExecuting(resolved, command)
	return commands.HandleCommand(resolved, cfg, systemPrompt.Text, newClient)
}

func run(args []string, newClient api.ClientFactory) error {
	set := flag.NewFlagSet("yact", flag.ContinueOnError)
	cli := config.RegisterFlags(set)
	if err := set.Parse(args); err != nil {
		return err
	}

	flags, err := cli.ToFlaglist(set)
	if err != nil {
		return err
	}

	if cli.Help {
		commands.ShowHelp()
		return nil
	}

	if cli.Version {
		fmt.Printf("yact %s\n", version)
		return nil
	}

	commandArgs := set.Args()

	if len(commandArgs) == 0 {
		return errors.New("no command provided. Run 'y --help' for usage information.")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	command := commandArgs[0]
	subArgs := []string{}
	if len(commandArgs) > 1 {
		subArgs = commandArgs[1:]
	}

	switch command {
	case "help":
		commands.ShowHelp()
		return nil
	case "read":
		return commands.HandleReadCommand(subArgs)
	case "keyword":
		if err := requireArgCount("keyword", subArgs, 1); err != nil {
			return err
		}
		return commands.HandleKeywordCommand(subArgs)
	case "config":
		if err := requireArgCount("config", subArgs, 0, 2); err != nil {
			return err
		}
		return commands.HandleConfigCommand(subArgs, cfg)
	case "new":
		return commands.HandleNewCommand()
	case "commit":
		if err := requireNoArgs("commit", subArgs); err != nil {
			return err
		}
		return commands.HandleCommitCommand()
	case "query":
		if err := requireNoArgs("query", subArgs); err != nil {
			return err
		}
		resolved := config.ResolveFlags(flags, config.QueryDefaults())
		printExecuting(resolved, "query")
		return commands.HandleCommand(resolved, cfg, "", newClient)
	case "batch":
		if err := requireNoArgs("batch", subArgs); err != nil {
			return err
		}
		return commands.HandleBatchCommand(cfg, newClient)
	case "download":
		return commands.HandleDownloadCommand(subArgs)
	case "index":
		if err := requireArgCount("index", subArgs, 1); err != nil {
			return err
		}
		return commands.HandleIndexCommand(subArgs)
	case "compact":
		if err := requireNoArgs("compact", subArgs); err != nil {
			return err
		}
		return commands.HandleCompactCommand()
	case "prompt":
		if err := requireArgCount("prompt", subArgs, 1); err != nil {
			return err
		}
		return commands.HandlePromptCommand(subArgs[0])
	default:
		if err := requireNoArgs(command, subArgs); err != nil {
			return err
		}
		return runPromptCommand(command, flags, cfg, newClient)
	}
}

func main() {
	if err := run(os.Args[1:], api.NewClaudeClient); err != nil {
		// Failing to write an error message leaves no other channel to report the failure,
		// so the process exits with the same status as the error it was trying to report.
		if _, writeErr := fmt.Fprintf(os.Stderr, "Error: %v\n", err); writeErr != nil {
			os.Exit(1)
		}
		os.Exit(1)
	}
}
