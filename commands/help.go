// Displays command reference and usage information
// Displays command reference and usage information
// Displays command reference and usage information
package commands

import (
	"fmt"
)

func ShowHelp() {
	fmt.Println(`yact - AI-powered code transformation tool

Usage:
  y [flags] <command> [arguments]

Commands:
  help                                 Show this help message
  read <file> [<file2> ...]            Add files to the task context (supports glob patterns)
  config [key] [value]                 Manage configuration settings
  act [prompt]                         Generate code modifications based on context
  snip <file> <start> <end> <prompt>   Modify specific line range in a file
  ask [prompt]                         Answer questions about code
  bash [prompt]                        Generate bash commands for a task
  plan [prompt]                        Create implementation plans
  context                              Discover relevant files for the current task
  index                                Create or update file index with descriptions
  new                                  Create a new empty task context
  buffer                               Output the content of the buffer log

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
  y -t act                             Generate code with thinking enabled
  y -o ask                             Answer question using Opus model
  y snip main.go 10 20 "fix this"      Modify lines 10-20 with prompt
  y --think --sonnet plan              Create plan using Sonnet with thinking
  y act "Is the implementation of Foo buggy?"  Set prompt and generate code

For more information, visit: https://github.com/yourusername/yact`)
}