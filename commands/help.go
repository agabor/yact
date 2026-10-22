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
  read <file> [<file2> ...]            Add files to the task context (supports glob patterns)
  message <text>                       Write the text to .yact/message.txt, sent before the index file, the output file and the context files
  output                               Read standard input into .yact/output.txt, sent after the index file and before the context files
  keyword <keyword>                    Recursively add files containing the keyword to the task context
  config [key] [value]                 Manage configuration settings
  new                                  Create a new empty task context
  query [prompt]                       Call the LLM without a system prompt
  commit                               Write code blocks from the buffer to disk (with -v, fail on free text or incomplete code blocks)
  index [private|public]               Recursively scan Go, PHP and JS files in the current directory and write .yact/index.txt with the files and their functions. private (default) indexes all functions and types, public indexes only the public interface
  <command> [prompt]                   Call the LLM with the systempront that belongs to the given command.

Prompt rules for query and <command>:
  A prompt argument is always used, together with the files in .yact/context.txt.
  Without a prompt argument, .yact/plan.md is used as the prompt and only the files
  linked in it are sent as context. A missing plan is an error.`)
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
  -d, --download                       Download the system prompt for the command
  --no-progress                        Hide progress indicator
  -v, --validate-code                  Fail if the response (or, for commit, the buffer) contains free text or incomplete code blocks
  -p, --plan                           Write the response to .yact/plan.md
  -s, --steps <selector>               Implement only the selected steps of .yact/plan.md. Requires .yact/plan.md and cannot be combined with a prompt argument or with --plan. The selector is N, N-M, a comma separated list such as 1,3,5-6, or next. Finished steps are ticked off in the plan after a successful write
  -m, --max-input-lines <n>            Set max input lines for the current command, without changing config
  -f, --file <path>                    Read the prompt from the given file instead of the command line argument

For more information, visit: https://github.com/agabor/yact`)
}

func ShowHelp() {
	printHelpHeader()
	printPromptCommands()
	printHelpFooter()
}