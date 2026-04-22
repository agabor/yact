package commands

import (
	"fmt"
)

func ShowHelp() {
	fmt.Println("y - Yet Another Coding Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  y act [prompt]          # Generate code with prompt")
	fmt.Println("  y bash [prompt]         # Generate a bash script file")
	fmt.Println("  y ask [question]        # Ask questions about the codebase")
	fmt.Println("  y plan [prompt]         # Get a plan for implementation")
	fmt.Println("  y step <index>          # Implement a specific step from the plan")
	fmt.Println("  y go                    # Execute the plan (alias for 'act Do it.')")
	fmt.Println("  y accept                # Accept last plan as user message")
	fmt.Println("  y read <file>           # Add file reference to prompt")
	fmt.Println("  y new                   # Create a new context")
	fmt.Println("  y last                  # Show last AI response")
	fmt.Println("  y config                # Show current configuration")
	fmt.Println("  y config <key> <value>  # Set configuration value")
	fmt.Println()
	fmt.Println("Configuration keys:")
	fmt.Println("  anthropic_api_key   Claude API key")
	fmt.Println("  claude_model        Claude model name")
}