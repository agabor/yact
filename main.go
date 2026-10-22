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

func readPromptFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", path, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		return "", fmt.Errorf("prompt file %s is empty", path)
	}
	return string(content), nil
}

func resolvePrompt(args []string, promptFile string) (string, error) {
	if promptFile == "" {
		return getOptionalPrompt(args), nil
	}
	if len(args) > 0 {
		return "", fmt.Errorf("--file cannot be combined with a prompt argument")
	}
	return readPromptFile(promptFile)
}

func getIndexVisibility(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "private"
}

func getModelOverride(claudeLevel int) string {
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

func hasUserSetModelFlag(claudeLevel int) bool {
	return claudeLevel > 0
}

func claudeLevelFromFlag(flag string) int {
	const fallbackClaudeLevel = 2
	if len(flag) < 2 {
		return fallbackClaudeLevel
	}
	digit := flag[1]
	if digit < '1' || digit > '4' {
		return fallbackClaudeLevel
	}
	return int(digit - '0')
}

func applyDefaultFlags(flags []string, skipModelFlags bool, noWriteFlag, downloadFlag, validateCodeFlag, planFlag *bool, claudeLevel *int) {
	for _, f := range flags {
		if strings.HasPrefix(f, "c") {
			if !skipModelFlags {
				*claudeLevel = claudeLevelFromFlag(f)
			}
			continue
		}
		switch f {
		case "n":
			if *validateCodeFlag == false {
				*noWriteFlag = true
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

func buildFlagString(noWriteFlag, downloadFlag, noProgressFlag, validateCodeFlag, planFlag bool, claudeLevel, maxInputLines int, steps string, promptFile string) string {
	var flags []string

	if noWriteFlag {
		flags = append(flags, "--no-write")
	}
	if claudeLevel > 0 {
		flags = append(flags, fmt.Sprintf("--claude %d", claudeLevel))
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
	if promptFile != "" {
		flags = append(flags, fmt.Sprintf("--file %s", promptFile))
	}

	return strings.Join(flags, " ")
}

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	noWriteFlag := flag.BoolP("no-write", "n", false, "Do not write files, print response instead")
	claudeLevel := flag.IntP("claude", "c", 0, "Claude model level (1: Haiku, 2: Sonnet, 3: Opus, 4: Fable)")
	downloadFlag := flag.BoolP("download", "d", false, "Download the system prompt for the command")
	noProgressFlag := flag.Bool("no-progress", false, "Hide progress indicator")
	validateCodeFlag := flag.BoolP("validate-code", "v", false, "Fail if the response contains free text or incomplete code blocks")
	planFlag := flag.BoolP("plan", "p", false, "Write the response to .yact/plan.md")
	maxInputLinesFlag := flag.IntP("max-input-lines", "m", 0, "Set max input lines for the current command, without changing config")
	stepsFlag := flag.StringP("steps", "s", "", "Select plan steps to implement: N, N-M, a comma separated list, or next")
	fileFlag := flag.StringP("file", "f", "", "Read the prompt from the given file instead of the command line argument")

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
		requireArgCount("message", commandArgs, 1)
		commandErr = commands.HandleMessageCommand(commandArgs)
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
		commandErr = commands.HandleCommitCommand(*validateCodeFlag)
	case "query":
		requireArgCount("query", commandArgs, 0, 1)
		prompt, promptErr := resolvePrompt(commandArgs, *fileFlag)
		if promptErr != nil {
			commandErr = promptErr
			break
		}
		modelOverride := getModelOverride(*claudeLevel)
		flagString := buildFlagString(*noWriteFlag, *downloadFlag, *noProgressFlag, *validateCodeFlag, *planFlag, *claudeLevel, *maxInputLinesFlag, *stepsFlag, *fileFlag)
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

		skipModelFlags := hasUserSetModelFlag(*claudeLevel)
		applyDefaultFlags(defaultFlags, skipModelFlags, noWriteFlag, downloadFlag, validateCodeFlag, planFlag, claudeLevel)

		prompt, promptErr := resolvePrompt(commandArgs, *fileFlag)
		if promptErr != nil {
			commandErr = promptErr
			break
		}

		cleanedPrompt := systemPrompt
		if strings.HasPrefix(cleanedPrompt, "flags: ") {
			lines := strings.Split(cleanedPrompt, "\n")
			if len(lines) > 1 {
				cleanedPrompt = strings.Join(lines[1:], "\n")
			}
		}

		modelOverride := getModelOverride(*claudeLevel)
		flagString := buildFlagString(*noWriteFlag, *downloadFlag, *noProgressFlag, *validateCodeFlag, *planFlag, *claudeLevel, *maxInputLinesFlag, *stepsFlag, *fileFlag)
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