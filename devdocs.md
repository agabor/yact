// Comprehensive code documentation for yact project
# YACT Developer Documentation

## Overview
YACT (Yet Another Coding Tool) is a CLI application that integrates with Claude AI to assist with code generation, analysis, and planning tasks. The codebase is organized into four main packages: `api`, `commands`, `config`, and `logic`.

## Package: api

### ClaudeClient (api/claude.go)
Struct that implements the Client interface for interacting with the Anthropic Claude API. Manages API credentials, model selection, token limits, and cost calculation.

**Methods:**
- `Init(cfg *config.Config)` - Initializes the client with configuration settings including API key, model selection, and token budgets
- `selectModel(cfg *config.Config) anthropic.Model` - Selects the appropriate Claude model variant (Opus, Sonnet, or Haiku) based on configuration
- `selectInputPrice(cfg *config.Config) int` - Returns the input token price in cents per million tokens for the selected model
- `selectOutputPrice(cfg *config.Config) int` - Returns the output token price in cents per million tokens for the selected model
- `GetModelName() string` - Returns the name of the currently selected model
- `calculateCost(inputTokens int64, outputTokens int64) float64` - Calculates API call cost based on token usage
- `Call(transaction logic.Transaction, think bool, systemPrompt string) (string, error)` - Sends a request to Claude with context files and prompts, returns the model's response

### Client Interface (api/client.go)
Interface defining the contract for AI client implementations with three required methods.

**Methods:**
- `Init(cfg *config.Config)` - Initializes the client with configuration
- `GetModelName() string` - Returns the model name as a string
- `Call(transaction logic.Transaction, think bool, systemPrompt string) (string, error)` - Executes an AI request and returns the response

### Serialize Function (api/client.go)
Reads a file from disk and formats it as a delimited code block with file path comment for inclusion in prompts.

## Package: commands

### HandleActCommand (commands/call.go)
Generates code based on the current transaction context and request, optionally writing generated files to disk or displaying response only.

### HandlePlanCommand (commands/call.go)
Requests a plan for implementation from Claude and saves it as the transaction request for later reference.

### HandleAskCommand (commands/call.go)
Sends a question about the codebase to Claude and displays the response without saving files.

### HandleConfigCommand (commands/config.go)
Manages user configuration by displaying current settings or setting individual configuration values like API key, model, and token limits.

### HandleContextCommand (commands/context.go)
Discovers relevant files for the current task by analyzing the task prompt against the project index and updates the transaction context accordingly.

### DiscoverRelevantFiles (commands/context.go)
Uses Claude to identify which files from a project index are most relevant to a given task prompt, returning a list of relevant file paths.

### BuildFileListingText (commands/context.go)
Constructs a formatted string containing all project files with their descriptions for presentation to Claude during context discovery.

### HandleIndexCommand (commands/index.go)
Scans the project directory, merges findings with existing index entries, fetches AI-generated descriptions for files without descriptions, and saves the updated index.

### GetFileDescriptions (commands/index.go)
Uses Claude to generate brief descriptions for files that lack descriptions in the project index.

### HandleNewCommand (commands/new.go)
Creates a new empty transaction context for starting a fresh task.

### HandleReadCommand (commands/read.go)
Adds file references to the current transaction context using glob pattern matching to support multiple files.

### HandleSnipCommand (commands/snip.go)
Requests code modifications for a specific line range in a file, applies Claude's response while preserving indentation, and writes the result back to the file.

### ShowHelp (commands/help.go)
Displays usage information for all available commands and configuration options.

## Package: config

### Config Struct (config/config.go)
Contains user configuration settings for API access and model behavior with fields for API key, model selection, and token budgets.

**Fields:**
- `AnthropicAPIKey string` - Claude API authentication key
- `ClaudeModel string` - Selected model variant (haiku, sonnet, or opus)
- `MaxTokens int` - Maximum output tokens per API call
- `ThinkBudget int` - Token budget for extended thinking mode

**Methods:**
- `Save() error` - Persists configuration to the user's home directory config file

### DefaultConfig Function (config/config.go)
Returns a Config struct with sensible default values for all fields.

### Load Function (config/config.go)
Reads configuration from the user's home directory, returning defaults if the config file doesn't exist.

### LoadPrompt Function (config/config.go)
Reads system prompt content from files stored in the user's config directory by prompt name.

## Package: logic

### CodeFile Struct (logic/codefile.go)
Represents a source code file with its path, content, and optional description extracted from code comments.

**Methods:**
- `Write() error` - Creates necessary directories and writes the code file to disk

### NewCodeFile Function (logic/codefile.go)
Creates a CodeFile from a path and content string, with automatic path normalization.

### ReadAsFile Function (logic/codefile.go)
Reads an existing file from disk and returns it as a CodeFile struct.

### ParseCodeFiles Function (logic/codefile.go)
Parses Claude's response to extract code blocks delimited by backticks, returning both CodeFile objects and any text outside code blocks.

### FileEntry Struct (logic/index.go)
Represents an indexed project file with its path and cached description.

**Fields:**
- `Path string` - Relative file path
- `Description string` - Brief description of file purpose

### LoadIndex Function (logic/index.go)
Reads the project file index from a CSV file, returning an empty index if the file doesn't exist.

### GetAllFiles Function (logic/index.go)
Recursively scans the project directory excluding common directories like .git and node_modules, returning entries for all discovered files.

### MergeIndex Function (logic/index.go)
Combines newly discovered disk files with existing index entries, preserving descriptions and removing entries for deleted files.

### SaveIndex Function (logic/index.go)
Writes the file index to a CSV file for later retrieval and analysis.

### Transaction Struct (logic/transaction.go)
Represents the context and request for an AI operation with two components: context files and the task request.

**Fields:**
- `Request []string` - The task description or prompt for Claude
- `Context []string` - Paths to files to include as context

**Methods:**
- `Save() error` - Persists the transaction to the .yact file in the current directory

### LoadTransaction Function (logic/transaction.go)
Reads the transaction from the .yact file in the current directory, returning an empty transaction if the file doesn't exist.