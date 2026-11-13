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
  keyword <keyword>                    Recursively add files containing the keyword to the task context
  config [key] [value]                 Manage configuration settings
  new                                  Create a new empty task context
  query [prompt]                       Call the LLM without a system prompt
  commit                               Write code blocks from the buffer to disk (with --strict, fail on free text or incomplete code blocks)
  batch                                Fetch the pending batch result and write it using the output flags it was submitted with (--output, --strict). Without them, the result goes to the buffer. Apply the buffer with commit
  download prompt <command>            Download a system prompt from the project repository, replacing the local copy
  index <none|public|all>              Write the index to .yact/temp/index.txt without calling the LLM, where none removes it, public is public and all is private
  compact                              Compact the links in the prompt file
  <command> [prompt]                   Call the LLM with the system prompt that belongs to the given command.

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
  -V, --version                        Print the version and exit
  -m, --model <name>                   Model: haiku, sonnet, opus, fable, or 1 to 4
  -o, --output <mode>                  apply writes the result to files, print prints it, plan writes it to .yact/plan.md. A plan sent as the prompt is sent whole, the same as --steps 0, and no step is ticked off
  --strict                             Fail if the response (or, for commit, the buffer) contains free text or incomplete code blocks. Only allowed with apply output
  -i, --index <name>                   Index for this call: none sends no index, public generates a public index, all generates a private index. The index is written to .yact/temp/index.txt and sent only in the call that sets the switch
  -s, --steps <selector>               Implement only the selected steps of .yact/plan.md. Requires .yact/plan.md and cannot be combined with a prompt argument or with --output plan. The selector is N, N-M, a comma separated list such as 1,3,5-6, next, or all. all sends the whole plan without step selection and ticks off no step. Finished steps are ticked off in the plan after a successful write
  -b, --batch                          Submit the request as a batch job, allows one pending job at a time, skips post-processing. The output flags (--output, --strict) are saved with the job and applied when the result is fetched with 'y batch'
  --no-progress                        Hide progress indicator
  --verbose                            Print all file paths and the prompt sent to the LLM
  --dry-run                            Skip the LLM call, only build and check the request

For more information, visit: https://github.com/agabor/yact`)
}

func ShowHelp() {
	printHelpHeader()
	printPromptCommands()
	printHelpFooter()
}
