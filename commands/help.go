package commands

import (
	"fmt"
)

func ShowHelp() {
	fmt.Println(`yact - Yet Another Coding Tool

Usage:
  y [flags] <command> [arguments]

Commands:
  help                                 Show this help message
  read <file> [<file2> ...]            Add files to the task context (supports glob patterns)
  config [key] [value]                 Manage configuration settings
  <command> [prompt]                   Call the LLM with the systempront that belongs to the given command.
  new                                  Create a new empty task context
  buffer                               Output the content of the buffer log
  stash                                Stash the current prompt
  stash pop                            Restore the most recently stashed prompt

Global Flags:
  -h, --help                           Show help message
  -t, --think                          Enable Claude's extended thinking mode
  -n, --no-write                       Do not write files, print response instead
  -f, --fable                          Use Claude Fable model
  -o, --opus                           Use Claude Opus model
  -s, --sonnet                         Use Claude Sonnet model
      --haiku                          Use Claude Haiku model

Examples:
  y --help                             Show this help message
  y config                             Show current configuration
  y config anthropic_api_key <key>     Set API key
  y read main.go "commands/*.go"       Add files to the task context
  y stash                              Stash the current prompt
  y stash pop                          Restore the most recently stashed prompt

For more information, visit: https://github.com/agabor/yact`)
}
