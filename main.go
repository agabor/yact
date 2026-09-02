package main

import (
	"fmt"
	"os"
	"strings"

	"yact/commands"
	"yact/config"
	"yact/logic"

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

func requireMinArgs(command string, args []string, min int) {
	if len(args) < min {
		fmt.Fprintf(os.Stderr, "Error: %s command requires at least %d argument(s)\n", command, min)
		os.Exit(1)
	}
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

func resolvePrompt(useBuffer bool, args []string) (string, error) {
	if useBuffer {
		bufferPrompt, err := logic.ReadBuffer()
		if err != nil {
			return "", err
		}
		return bufferPrompt, nil
	}
	return getOptionalPrompt(args), nil
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

func applyDefaultFlags(flags []string, skipModelFlags bool, noWriteFlag, bufferFlag, downloadFlag, validateCodeFlag, noContextFlag, noSavePromptFlag *bool, claudeLevel, qwenLevel *int) {
	for _, f := range flags {
		switch f {
		case "n":
			*noWriteFlag = true
		case "c":
			if !skipModelFlags {
				*claudeLevel = 2
			}
		case "q":
			if !skipModelFlags {
				*qwenLevel = 2
			}
		case "b":
			*bufferFlag = true
		case "d":
			*downloadFlag = true
		case "v":
			*validateCodeFlag = true
		case "x":
			*noContextFlag = true
		case "p":
			*noSavePromptFlag = true
		}
	}
}

func buildFlagString(noWriteFlag, bufferFlag, downloadFlag, noProgressFlag, validateCodeFlag, noContextFlag, noSavePromptFlag bool, claudeLevel, qwenLevel int) string {
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
	if bufferFlag {
		flags = append(flags, "--buffer")
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
	if noContextFlag {
		flags = append(flags, "--no-context")
	}
	if noSavePromptFlag {
		flags = append(flags, "--no-save-prompt")
	}

	return strings.Join(flags, " ")
}

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	noWriteFlag := flag.BoolP("no-write", "n", false, "Do not write files, print response instead")
	claudeLevel := flag.IntP("claude", "c", 0, "Claude model level (1: Haiku, 2: Sonnet, 3: Opus, 4: Fable)")
	qwenLevel := flag.IntP("qwen", "q", 0, "Qwen model level (1: Coder, 2: Max)")
	bufferFlag := flag.BoolP("buffer", "b", false, "Use the buffer content as the prompt")
	downloadFlag := flag.BoolP("download", "d", false, "Download the system prompt for the command")
	noProgressFlag := flag.Bool("no-progress", false, "Hide progress indicator")
	validateCodeFlag := flag.BoolP("validate-code", "v", false, "Fail if the response contains free text or incomplete code blocks")
	noContextFlag := flag.BoolP("no-context", "x", false, "Do not send the selected files to the LLM")
	noSavePromptFlag := flag.BoolP("no-save-prompt", "p", false, "Do not update prompt.txt with the prompt given as a CLI argument")

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
	case "keyword":
		requireArgCount("keyword", commandArgs, 1)
		commandErr = commands.HandleKeywordCommand(commandArgs)
	case "tag":
		requireMinArgs("tag", commandArgs, 2)
		commandErr = commands.HandleTagCommand(commandArgs)
	case "config":
		requireArgCount("config", commandArgs, 0, 2)
		commandErr = commands.HandleConfigCommand(commandArgs, cfg)
	case "prompt":
		if *bufferFlag {
			requireArgCount("prompt", commandArgs, 0)
		} else {
			requireArgCount("prompt", commandArgs, 1)
		}
		commandErr = commands.HandlePromptCommand(commandArgs, *bufferFlag)
	case "new":
		commandErr = commands.HandleNewCommand()
	case "query":
		if *bufferFlag {
			requireArgCount("query", commandArgs, 0)
		} else {
			requireArgCount("query", commandArgs, 0, 1)
		}
		prompt, promptErr := resolvePrompt(*bufferFlag, commandArgs)
		if promptErr != nil {
			commandErr = promptErr
			break
		}
		modelOverride := getModelOverride(*claudeLevel, *qwenLevel)
		flagString := buildFlagString(*noWriteFlag, *bufferFlag, *downloadFlag, *noProgressFlag, *validateCodeFlag, *noContextFlag, *noSavePromptFlag, *claudeLevel, *qwenLevel)
		if flagString != "" {
			fmt.Printf("Executing: y %s query\n", flagString)
		}
		commandErr = commands.HandleCommand(true, *noProgressFlag, *validateCodeFlag, *noContextFlag, *noSavePromptFlag, cfg, "", modelOverride, prompt)
	default:
		if *bufferFlag {
			requireArgCount(command, commandArgs, 0)
		} else {
			requireArgCount(command, commandArgs, 0, 1)
		}
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
		applyDefaultFlags(defaultFlags, skipModelFlags, noWriteFlag, bufferFlag, downloadFlag, validateCodeFlag, noContextFlag, noSavePromptFlag, claudeLevel, qwenLevel)

		prompt, promptErr := resolvePrompt(*bufferFlag, commandArgs)
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

		modelOverride := getModelOverride(*claudeLevel, *qwenLevel)
		flagString := buildFlagString(*noWriteFlag, *bufferFlag, *downloadFlag, *noProgressFlag, *validateCodeFlag, *noContextFlag, *noSavePromptFlag, *claudeLevel, *qwenLevel)
		if flagString != "" {
			fmt.Printf("Executing: y %s %s\n", flagString, command)
		}
		commandErr = commands.HandleCommand(*noWriteFlag, *noProgressFlag, *validateCodeFlag, *noContextFlag, *noSavePromptFlag, cfg, cleanedPrompt, modelOverride, prompt)
	}

	if commandErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		os.Exit(1)
	}
}