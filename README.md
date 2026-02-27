# YACT - Yet Another Coding Tool

Minimal, fast, transactional LLM coding assistant

## RAG vs Prompt Stuffing

## Conversational vs Transactional

Most mainstream coding assistants, like Cline, Cursor and Copilot, are conversational. They maintain a messaging history,
on a chat interface. They are proactive, capable of reading files by themselves, making modifications, and testing those
modifications.

YACT, on the other hand, is transactional. A YACT transaction is a request response pair. The request comes from the user,
and contains a list of file contents, with a textual prompt. The response comes from the LLM, and either contains a list
of file contents or some textual description.

## Tool Usage vs Code Blocks

Most coding agents use a set of pre-defined tools to work with your code base. LLMs use XML formated messages, to signal
tool usage intent. The agent parses the response, and executes the specified tools.
For example, Cline's tools are the following:
https://docs.cline.bot/tools-reference/all-cline-tools

| Tool	                        | Description                     |
|------------------------------|---------------------------------|
| `write_to_file`              | 	Create or overwrite files      |
| `read_file`                  | 	Read file contents             |
| `replace_in_file`	           | Make targeted edits to files    |
| `search_files`	              | Search files using regex        |
| `list_files`	                | List directory contents         |
| `list_code_definition_names` | 	List code definitions in files |

YACT, on the other hand, does not have any tools, but uses Markdown formated code blocks.

## Comparison
|                          | Conversational                                                                                 | Transactional                                                                                                                                                              |
|--------------------------|------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Keeps messaging history  | Yes                                                                                            | No, most previous messages are pruned.                                                                                                                                     |
| Generation method        | Retrieval Augumented Generation (RAG)                                                          | Prompt Stuffing                                                                                                                                                            |
| LLM interface complexity | Complex                                                                                        | Simple                                                                                                                                                                     |
| Model Selection          | [Larger Models](https://docs.cline.bot/running-models-locally/overview#why-not-smaller-models) | Works with larger a selection of models, down to 7b parameters.                                                                                                            |
| MCP usage                | Yes                                                                                            | No                                                                                                                                                                         |
| API Cost                 | Up to $50 per day                                                                              | Up to $2 per day                                                                                                                                                           |
| Vibe Coding Capable      | Yes                                                                                            | No                                                                                                                                                                         |
| File size limitations    | No limitation, however, they tend to be less effective on large source files.                  | Editable file sizes are limited by the LLMs output token limit. This typically means a maximum of 400 LOC per file, however keeping file sizes under 200 LOC is recomended |
| Coding standards         | No official recomendations.                                                                    | Strict clean coding principles are recomended.                                                                                                                             |

## Installation

Download the latest binary for your operating system from the [GitHub releases page](https://github.com/agabor/yact/releases).

Available binaries:
- Linux: `yact-linux-amd64`
- Windows: `yact-windows-amd64.exe`
- macOS: `yact-darwin-amd64` or `yact-darwin-arm64`

For example, to download and install on Linux:

```bash
cd /tmp
wget https://github.com/agabor/yact/releases/download/v0.0.0/yact-linux-amd64
chmod +x yact-linux-amd64
sudo mv yact-linux-amd64 /usr/local/bin/y
```

On macOS:

```bash
cd /tmp
curl -L -o yact-darwin-amd64 https://github.com/agabor/yact/releases/download/v0.0.0/yact-darwin-amd64
chmod +x yact-darwin-amd64
sudo mv yact-darwin-amd64 /usr/local/bin/y
```

On Windows, download the `yact-windows-amd64.exe` file and place it in a directory that's in your PATH, or run it directly from where you download it.

## Quick Start

Before using `y`, configure your Claude API key:

```bash
y config anthropic_api_key YOUR_API_KEY
```

You can verify your configuration at any time:

```bash
y config
```

## Commands

### Generate Code

Generate code files based on a prompt:

```bash
y act "create a user authentication handler"
```

The AI will generate complete files with code blocks. By default, files are written directly to your filesystem. Use the `--safe` flag to add a `.new` suffix to generated files for review before replacing originals:

```bash
y act --safe "add logging to the user service"
```

### Generate Bash Scripts

Generate standalone bash scripts:

```bash
y bash "create a script that backs up my database"
```

### Ask Questions

Ask questions about your codebase or general topics:

```bash
y ask "how does the authentication flow work?"
y ask "what are the best practices for Go error handling?"
```

### File References

Attach files to your prompts so the AI can reference them when generating code:

```bash
y read src/models/user.go
y read src/handlers/*.go
y ask "add validation to the user model"
```

Use glob patterns to match multiple files. View your current attachments:

```bash
y list
```

Clear all attachments:

```bash
y clear
```

### Context Management

By default, `y` maintains a conversation history. Use this to build on previous responses:

```bash
y ask "what functions do we need?"
y act "now implement those functions"  # The AI remembers the previous question
```

Start a fresh conversation:

```bash
y new
```

Retrieve the last AI response:

```bash
y last
```

## Configuration

View current settings:

```bash
y config
```

Set configuration values:

```bash
y config anthropic_api_key your_key_here
y config claude_model claude-opus-4-1-20250805
```

Available configuration keys:
- `anthropic_api_key` - Your Claude API key (required)
- `claude_model` - Which Claude model to use (default: claude-haiku-4-5-20251001)

## Piping Input

You can pipe text directly to `y`:

```bash
cat requirements.txt | y act
echo "fix the database connection" | y bash
```

## Usage Tips

- **File patterns**: Use glob patterns with `read` to attach multiple related files at once
- **Safe mode**: Always use `--safe` when generating new code to review changes first
- **Context**: Build up context by reading relevant files before asking complex questions
- **Conversation flow**: Use multi-step conversations - ask for a plan first, then generate code
- **Cost tracking**: The tool displays token usage and estimated costs for each API call

## Workflow

The following diagram illustrates YACT message types as nodes YACT commands as edges:

```mermaid
graph TD
    File -->|ask| Question
    File -->|act| Command
    File -->|plan| Objective
    Question -->|assistant| Answer
    Answer -->|ask| Question
    Command -->|assistant| Action
    Action -->|act| Command
    Objective -->|assistant| Plan
    Plan -->|act| Command
```

## Help

Display command reference:

```bash
y help
y --help
y -h
```

## Storage

Configuration and conversation history are stored in `~/.yact/`:
- `config` - Your API key and model settings
- `context.json` - Conversation history
- `attachments.json` - List of attached files

## Troubleshooting

**"Claude API key not configured"**
- Set your API key: `y config anthropic_api_key YOUR_KEY`

**"No files found matching pattern"**
- Check that the glob pattern matches existing files
- Use exact paths if glob patterns don't work

**"File not found in attachments"**
- Verify the file path is correct and the file exists
- Use `y list` to see currently attached files

**API errors**
- Verify your API key is valid
- Check your internet connection
- Ensure your Claude API account has available credits