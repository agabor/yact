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

func getIndexVisibility(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "private"
}

func getModelOverride(claudeLevel, qwenLevel int) string {
	if claudeLevel > 0 {
		switch claudeLevel {
		case 1:
			return "haiku"
		case 2:
			return "sonnet"
		case 3:
			return "opus"
		case 4:
			return "fable"
		}
	}
	if qwenLevel > 0 {
		switch qwenLevel {
		case 1:
			return "qnext"
		case 2:
			return "qmax"
		}
	}
	return ""
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

func hasUserSetModelFlag(claudeLevel, qwenLevel int) bool {
	return claudeLevel > 0 || qwenLevel > 0
}

func applyDefaultFlags(flags []string, skipModelFlags bool, noWriteFlag, downloadFlag, validateCodeFlag, planFlag *bool, claudeLevel, qwenLevel *int) {
	for _, f := range flags {
		switch f {
		case "n":
			if *validateCodeFlag == false {
				*noWriteFlag = true
			}
		case "c":
			if !skipModelFlags {
				*claudeLevel = 2
			}
		case "q":
			if !skipModelFlags {
				*qwenLevel = 2
			}
		case "d":
			*downloadFlag = true
		case "v":
			if *noWriteFlag == false {
				*validateCodeFlag = true
			}
		case "p":
			*planFlag = true
		}
	}
}

func buildFlagString(noWriteFlag, downloadFlag, noProgressFlag, validateCodeFlag, planFlag bool, claudeLevel, qwenLevel, maxInputLines int, steps string) string {
	var flags []string

	if noWriteFlag {
		flags = append(flags, "--no-write")
	}
	if claudeLevel > 0 {
		flags = append(flags, fmt.Sprintf("--claude %d", claudeLevel))
	}
	if qwenLevel > 0 {
		flags = append(flags, fmt.Sprintf("--qwen %d", qwenLevel))
	}
	if downloadFlag {
		flags = append(flags, "--download")
	}
	if noProgressFlag {
		flags = append(flags, "--no-progress")
	}
	if validateCodeFlag {
		flags = append(flags, "--validate-code")
	}
	if planFlag {
		flags = append(flags, "--plan")
	}
	if maxInputLines > 0 {
		flags = append(flags, fmt.Sprintf("--max-input-lines %d", maxInputLines))
	}
	if steps != "" {
		flags = append(flags, fmt.Sprintf("--steps %s", steps))
	}

	return strings.Join(flags, " ")
}

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	noWriteFlag := flag.BoolP("no-write", "n", false, "Do not write files, print response instead")
	claudeLevel := flag.IntP("claude", "c", 0, "Claude model level (1: Haiku, 2: Sonnet, 3: Opus, 4: Fable)")
	qwenLevel := flag.IntP("qwen", "q", 0, "Qwen model level (1: Coder, 2: Max)")
	downloadFlag := flag.BoolP("download", "d", false, "Download the system prompt for the command")
	noProgressFlag := flag.Bool("no-progress", false, "Hide progress indicator")
	validateCodeFlag := flag.BoolP("validate-code", "v", false, "Fail if the response contains free text or incomplete code blocks")
	planFlag := flag.BoolP("plan", "p", false, "Write the response to .yact/plan.md")
	maxInputLinesFlag := flag.IntP("max-input-lines", "m", 0, "Set max input lines for the current command, without changing config")
	stepsFlag := flag.StringP("steps", "s", "", "Select plan steps to implement: N, N-M, a comma separated list, or next")

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
	case "message":
		commandErr = commands.HandleMessageCommand(commandArgs)
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
		commandErr = commands.HandleCommitCommand(*validateCodeFlag)
	case "query":
		requireArgCount("query", commandArgs, 0, 1)
		prompt := getOptionalPrompt(commandArgs)
		modelOverride := getModelOverride(*claudeLevel, *qwenLevel)
		flagString := buildFlagString(*noWriteFlag, *downloadFlag, *noProgressFlag, *validateCodeFlag, *planFlag, *claudeLevel, *qwenLevel, *maxInputLinesFlag, *stepsFlag)
		if flagString != "" {
			fmt.Printf("Executing: y %s query\n", flagString)
		}
		commandErr = commands.HandleCommand(true, *noProgressFlag, *validateCodeFlag, *planFlag, cfg, "", modelOverride, *maxInputLinesFlag, *stepsFlag, prompt)
	case "index":
		requireArgCount("index", commandArgs, 0, 1)
		commandErr = commands.HandleIndexCommand(getIndexVisibility(commandArgs))
	default:
		requireArgCount(command, commandArgs, 0, 1)
		downloaded := *downloadFlag
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

		skipModelFlags := hasUserSetModelFlag(*claudeLevel, *qwenLevel)
		applyDefaultFlags(defaultFlags, skipModelFlags, noWriteFlag, downloadFlag, validateCodeFlag, planFlag, claudeLevel, qwenLevel)

		prompt := getOptionalPrompt(commandArgs)

		cleanedPrompt := systemPrompt
		if strings.HasPrefix(cleanedPrompt, "flags: ") {
			lines := strings.Split(cleanedPrompt, "\n")
			if len(lines) > 1 {
				cleanedPrompt = strings.Join(lines[1:], "\n")
			}
		}

		modelOverride := getModelOverride(*claudeLevel, *qwenLevel)
		flagString := buildFlagString(*noWriteFlag, *downloadFlag, *noProgressFlag, *validateCodeFlag, *planFlag, *claudeLevel, *qwenLevel, *maxInputLinesFlag, *stepsFlag)
		if flagString != "" {
			fmt.Printf("Executing: y %s %s\n", flagString, command)
		}
		commandErr = commands.HandleCommand(*noWriteFlag, *noProgressFlag, *validateCodeFlag, *planFlag, cfg, cleanedPrompt, modelOverride, *maxInputLinesFlag, *stepsFlag, prompt)
	}

	if commandErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		os.Exit(1)
	}
}