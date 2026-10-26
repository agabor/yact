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
  output                               Read standard input into .yact/output.txt, sent after the index file and before the context files
  keyword <keyword>                    Recursively add files containing the keyword to the task context
  config [key] [value]                 Manage configuration settings
  new                                  Create a new empty task context
  query [prompt]                       Call the LLM without a system prompt
  commit                               Write code blocks from the buffer to disk (with -v, fail on free text or incomplete code blocks)
  batch                                Fetch the pending batch result and write it using the output flags it was submitted with (-n, -v, -p). Without them, the result goes to the buffer. Apply the buffer with commit
  index <level>                        Write the index to .yact/index.txt without calling the LLM, where 0 removes it, 1 is public and 2 is private
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
  -i, --index <level>                  Index for this call: 0 sends no index, 1 generates a public index, 2 generates a private index. The index is written to .yact/index.txt and sent only in the call that sets the switch
  -b, --batch                          Submit the request as a batch job, allows one pending job at a time, skips post-processing. The output flags (-n, -v, -p) are saved with the job and applied when the result is fetched with 'y batch'
  -d, --download                       Download the system prompt for the command
  --no-progress                        Hide progress indicator
  -v, --validate-code                  Fail if the response (or, for commit, the buffer) contains free text or incomplete code blocks
  -p, --plan                           Write the response to .yact/plan.md
  -s, --steps <selector>               Implement only the selected steps of .yact/plan.md. Requires .yact/plan.md and cannot be combined with a prompt argument or with --plan. The selector is N, N-M, a comma separated list such as 1,3,5-6, or next. Finished steps are ticked off in the plan after a successful write
  -m, --max-input-lines <n>            Set max input lines for the current command, without changing config

For more information, visit: https://github.com/agabor/yact`)
}

func ShowHelp() {
	printHelpHeader()
	printPromptCommands()
	printHelpFooter()
}
