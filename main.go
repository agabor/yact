package main

import (
	"fmt"
	"os"

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

func getModelOverride(fable, opus, sonnet, haiku, qmax, qcoder bool) string {
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
	if qcoder {
		return "qcoder"
	}
	if qmax {
		return "qmax"
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

func main() {
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	thinkFlag := flag.BoolP("think", "t", false, "Enable Claude's extended thinking mode")
	noWriteFlag := flag.BoolP("no-write", "n", false, "Do not write files, print response instead")
	fableFlag := flag.BoolP("fable", "f", false, "Use Claude Fable model")
	opusFlag := flag.BoolP("opus", "o", false, "Use Claude Opus model")
	sonnetFlag := flag.BoolP("sonnet", "s", false, "Use Claude Sonnet model")
	haikuFlag := flag.Bool("haiku", false, "Use Claude Haiku model")
	qwenFlag := flag.BoolP("qmax", "w", false, "Use Qwen3 235B A22B Instruct 2507 model on AWS Bedrock")
	coderFlag := flag.BoolP("qcoder", "e", false, "Use Qwen3 Coder 30B A3B model (AWS Bedrock)")
	bufferFlag := flag.BoolP("buffer", "b", false, "Use the buffer content as the prompt")
	downloadFlag := flag.BoolP("download", "d", false, "Download the system prompt for the command")
	quietFlag := flag.BoolP("quiet", "q", false, "Hide progress indicator")
	codeOnlyFlag := flag.BoolP("code-only", "c", false, "Only accept code blocks in the response")
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

	modelOverride := getModelOverride(*fableFlag, *opusFlag, *sonnetFlag, *haikuFlag, *qwenFlag, *coderFlag)

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
		commandErr = commands.HandleCommand(*thinkFlag, true, *quietFlag, *codeOnlyFlag, *noContextFlag, *noSavePromptFlag, cfg, "", modelOverride, prompt)
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
		prompt, promptErr := resolvePrompt(*bufferFlag, commandArgs)
		if promptErr != nil {
			commandErr = promptErr
			break
		}
		systemPrompt, err := config.LoadPrompt(command)
		if err != nil {
			promptNotFoundMessage(command, downloaded)
			os.Exit(1)
		}
		commandErr = commands.HandleCommand(*thinkFlag, *noWriteFlag, *quietFlag, *codeOnlyFlag, *noContextFlag, *noSavePromptFlag, cfg, systemPrompt, modelOverride, prompt)
	}

	if commandErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", commandErr)
		os.Exit(1)
	}
}