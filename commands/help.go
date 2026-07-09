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
  read <file>                          Read and display file contents
  config [key] [value]                 Manage configuration settings
  act                                  Generate code modifications based on context
  snip <file> <start> <end> <prompt>   Modify specific code snippet
  ask                                  Answer questions about code
  plan                                 Create implementation plans
  context                              Display current context files
  index                                Create or update file index
  new                                  Create new project

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
  y config anthropic_api_key <key>    Set API key
  y -t act                             Generate code with thinking enabled
  y -o ask                             Answer question using Opus model
  y snip main.go 10 20 "fix this"     Modify lines 10-20 with prompt
  y --think --sonnet plan              Create plan using Sonnet with thinking

For more information, visit: https://github.com/yourusername/yact`)
}