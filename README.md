# YACT - Yet Another Coding Tool

Minimal, responsive, transparent LLM coding assistant

## Design principles

 * YACT supports AI assisted software development. (Not AI driven software development.)
 * YACT is optimized for responsiveness. You get results in seconds, not minutes.
 * YACT is reactive, not proactive. It does what it is explicitly asked to do, and nothing more.
 * YACT works transparently. You can read all messages sent to, and received from the LLM.

## Comparison to Mainstream Coding Agents

YACT differs from mainstream coding assistants, like Cline, Cursor and Copilot, in the sense that it not aims to be a
complete, automated coding suite, rather be a Swiss army knife for LLM assisted coding. This is why it is called a
"Coding Tool" rather than a "Coding Agent".

### RAG vs Prompt Stuffing

Mainstream coding agents typically use the Retrieval Augmented Generation (RAG) approach. This means, that the LLM
decides what files it needs to read to complete the given task, and uses the agent's file reading tool to retrieve
to necessary file contents.

YACT uses the prompt stuffing approach. The developer uses YACT's explicit read command, to append the content of a
file to the prompt. The LLM can not initiate file reading.

### Proactive vs Reactive

Mainstream coding agents typically utilize a chat interface. Once the developer sends an initial prompt, they are free to
proactively explore the codebase, run commands, edit files (with permission), while sending multiple API calls to the LLM.

YACT, on the other hand, is reactive. The developer needs to select the files needed for the task, and explicitly append
them to the prompt with the read command. Once the prompt is ready, a single API call will be made to the LLM. Depending
on the task,the LLM will respond in human language, or with generated code, that will be inserted into the code base by
YACT.

### Tool Usage vs Code Blocks

The earliest LLMs that were available through API had a really strict output token limit, as low as 4000 tokens, which
meant, that only 100 to 200 lines of codes could be generated in a single API call. To overcome this restriction agent
developers introduced file editing tool, to enable their agents to incrementally edit code files.

Today, typical output token limits are around 64K to 128K, which, with adequate clean coding practices, is plenty 
to generate multiple complete source files. Exploiting this "late comer's advantage", YACT always generates
complete source files, making its LLM interface far simpler than most agents. This approach also enables YACT to
aggressively compact its context, keeping API calls responsive, and cheap.

### Comparison
|                          | Cline / Cursor / Copilot                                                                       | YACT                                                                                                                              |
|--------------------------|------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| Interface Type           | Chat                                                                                           | CLI                                                                                                                               |
| Generation method        | Retrieval Augumented Generation (RAG)                                                          | Prompt Stuffing                                                                                                                   |
| LLM interface complexity | Complex                                                                                        | Simple                                                                                                                            |
| Model Selection          | [Larger Models](https://docs.cline.bot/running-models-locally/overview#why-not-smaller-models) | Works with larger a selection of models, down to 7b parameters.                                                                   |
| MCP usage                | Yes                                                                                            | No                                                                                                                                |
| API Cost                 | Up to $50 per day                                                                              | Up to $2 per day                                                                                                                  |
| Vibe Coding Capable      | Yes                                                                                            | No                                                                                                                                |
| File size limitations    | No limitation, however, they tend to be less effective on large source files.                  | Editable file sizes are limited by the LLMs output token limit. Limiting file sizes to a few hundres lines of code is recomended. |
| Coding standards         | No official recomendations.                                                                    | Strict clean coding principles are recomended.                                                                                    |

## Modes
YACT uses Question / Plan / Act modes. Then plan and act mode can be familiar for users of competing tools. The plan mode creates a step-by-step implementation plan for a given task, which can be implemented using act mode. In the question mode you can ask questions about you code base without interfering with coding tasks. 


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
