package commands

import (
	"fmt"

	"yact/config"
)

func printHelpHeader() {
	fmt.Println(`yact - Yet Another Coding Tool

Usage:
  y [flags] <command> [arguments]

Commands:
  help                                 Show this help message
  read <file> [<file2> ...]            Add files to the task context (supports glob patterns and tags)
  keyword <keyword>                    Recursively add files containing the keyword to the task context
  tag <tagname> <file> [<file2> ...]   Tag files (supports glob patterns) for later use with the read command
  config [key] [value]                 Manage configuration settings
  new                                  Create a new empty task context
  query [prompt]                       Call the LLM without a system prompt
  <command> [prompt]                   Call the LLM with the systempront that belongs to the given command.`)
}

func printPromptCommands() {
	promptNames, err := config.ListPromptNames()
	if err != nil || len(promptNames) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("  Available Commands:")
	for _, name := range promptNames {
		fmt.Printf("  %s\n", name)
	}
}

func printHelpFooter() {
	fmt.Println(`
Global Flags:
  -h, --help                           Show help message
  -n, --no-write                       Do not write files, print response instead
  -c, --claude <level>                 Use Claude model (1: Haiku, 2: Sonnet, 3: Opus, 4: Fable)
  -q, --qwen <level>                   Use Qwen model (1: Coder, 2: Max)
  -b, --buffer                         Use the buffer content as the prompt
  -d, --download                       Download the system prompt for the command
  --no-progress                        Hide progress indicator
  -v, --validate-code                  Fail if the response contains free text or incomplete code blocks
  -x, --no-context                     Do not send the selected files to the LLM
  -p, --no-save-prompt                 Do not update prompt.txt with the prompt given as a CLI argument

For more information, visit: https://github.com/agabor/yact`)
}

func ShowHelp() {
	printHelpHeader()
	printPromptCommands()
	printHelpFooter()
}