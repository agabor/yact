package main

import (
	"fmt"
	"os"
	"strings"

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

func getOptionalPrompt(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

func validateIndexLevel(level int) error {
	switch level {
	case 0, 1, 2:
		return nil
	}
	return fmt.Errorf("invalid index level %d, expected 0, 1 or 2", level)
}

func promptNotFoundMessage(command string, downloaded bool) {
	fmt.Fprintf(os.Stderr, "Error: no system prompt found for command '%s'\n", command)
	if !downloaded {
		fmt.Fprintf(os.Stderr, "Try downloading it with: y -d %s\n", command)
	} else {
		fmt.Fprintf(os.Stderr, "Download failed or no prompt exists for '%s'.\n", command)
	}
	fmt.Fprintln(os.Stderr, "Run 'y --help' for usage information.")
}

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	noWriteFlag := flag.BoolP("no-write", "n", false, "Do not write files, print response instead")
	claudeLevel := flag.IntP("claude", "c", 0, "Claude model level (1: Haiku, 2: Sonnet, 3: Opus, 4: Fable)")
	downloadFlag := flag.BoolP("download", "d", false, "Download the system prompt for the command")
	noProgressFlag := flag.Bool("no-progress", false, "Hide progress indicator")
	validateCodeFlag := flag.BoolP("validate-code", "v", false, "Fail if the response contains free text or incomplete code blocks")
	planFlag := flag.BoolP("plan", "p", false, "Write the response to .yact/plan.md")
	batchFlag := flag.BoolP("batch", "b", false, "Submit the request as a batch job instead of streaming it")
	maxInputLinesFlag := flag.IntP("max-input-lines", "m", 0, "Set max input lines for the current command, without changing config")
	stepsFlag := flag.StringP("steps", "s", "", "Select plan steps to implement: N, N-M, a comma separated list, or next")
	indexFlag := flag.IntP("index", "i", 0, "Index level for this call: 0 none, 1 public, 2 private")
	verboseFlag := flag.Bool("verbose", false, "Print all file paths and the prompt sent to the LLM")
	dryFlag := flag.Bool("dry", false, "Skip the LLM call, only build and check the request")

	flag.Parse()

	if err := validateIndexLevel(*indexFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()

	flags := config.Flaglist{
		NoWrite:       *noWriteFlag,
		NoProgress:    *noProgressFlag,
		ValidateCode:  *validateCodeFlag,
		WritePlan:     *planFlag,
		Batch:         *batchFlag,
		Download:      *downloadFlag,
		Verbose:       *verboseFlag,
		Dry:           *dryFlag,
		ClaudeLevel:   *claudeLevel,
		MaxInputLines: *maxInputLinesFlag,
		Steps:         *stepsFlag,
		IndexLevel:    *indexFlag,
	}

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
	case "output":
		requireNoArgs("output", commandArgs)
		commandErr = commands.HandleOutputCommand(commandArgs)
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
		commandErr = commands.HandleCommitCommand(flags.ValidateCode)
	case "query":
		requireArgCount("query", commandArgs, 0, 1)
		if flagString := flags.FlagString(); flagString != "" {
			fmt.Printf("Executing: y %s query\n", flagString)
		}
		queryFlags := flags
		queryFlags.NoWrite = true
		commandErr = commands.HandleCommand(queryFlags, cfg, "", getOptionalPrompt(commandArgs))
	case "batch":
		requireNoArgs("batch", commandArgs)
		commandErr = commands.HandleBatchCommand(cfg)
	case "index":
		requireArgCount("index", commandArgs, 1)
		commandErr = commands.HandleIndexCommand(commandArgs)
	default:
		requireArgCount(command, commandArgs, 0, 1)
		downloaded := flags.Download
		if downloaded {
			if err := config.DownloadPrompt(command); err != nil {
				commandErr = err
				break
			}
		}

		systemPrompt, defaultFlags, err := config.LoadPromptWithFlags(command)
		if err != nil {
			promptNotFoundMessage(command, downloaded)
			os.Exit(1)
		}

		flags.ApplyDefaults(defaultFlags, flags.HasUserSetModel())

		cleanedPrompt := systemPrompt
		if strings.HasPrefix(cleanedPrompt, "flags: ") {
			lines := strings.Split(cleanedPrompt, "\n")
			if len(lines) > 1 {
				cleanedPrompt = strings.Join(lines[1:], "\n")
			}
		}

		if flagString := flags.FlagString(); flagString != "" {
			fmt.Printf("Executing: y %s %s\n", flagString, command)
		}
		commandErr = commands.HandleCommand(flags, cfg, cleanedPrompt, getOptionalPrompt(commandArgs))
	}

	if commandErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		os.Exit(1)
	}
}
