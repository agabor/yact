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
the necessary file contents.

YACT uses the prompt stuffing approach. The developer uses YACT's explicit read command to add files to the task
context. The content of these files is appended to the prompt. The LLM can not initiate file reading.

### Proactive vs Reactive

Mainstream coding agents typically utilize a chat interface. Once the developer sends an initial prompt, they are free to
proactively explore the codebase, run commands, edit files (with permission), while sending multiple API calls to the LLM.

YACT, on the other hand, is reactive. The developer needs to select the files needed for the task, and explicitly append
them to the prompt with the read command. Once the prompt is ready, a single API call will be made to the LLM. Depending
on the task, the LLM will respond in human language, or with generated code, that will be inserted into the code base by
YACT.

### Tool Usage vs Code Blocks

The earliest LLMs that were available through API had a really strict output token limit, as low as 4000 tokens, which
meant, that only 100 to 200 lines of codes could be generated in a single API call. To overcome this restriction agent
developers introduced file editing tools, to enable their agents to incrementally edit code files.

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
| Model Selection          | [Larger Models](https://docs.cline.bot/running-models-locally/overview#why-not-smaller-models) | Works with a larger selection of models, down to 7b parameters.                                                                   |
| MCP usage                | Yes                                                                                            | No                                                                                                                                |
| API Cost                 | Up to $50 per day                                                                              | Up to $2 per day                                                                                                                  |
| Vibe Coding Capable      | Yes                                                                                            | No                                                                                                                                |
| File size limitations    | No limitation, however, they tend to be less effective on large source files.                  | Editable file sizes are limited by the LLMs output token limit. Limiting file sizes to a few hundred lines of code is recomended. |
| Coding standards         | No official recomendations.                                                                    | Strict clean coding principles are recomended.                                                                                    |

## Modes

YACT uses Ask / Plan / Act modes. The plan and act modes can be familiar for users of competing tools. The plan mode
creates a step-by-step implementation plan for a given task, which can then be implemented using act mode. In ask mode
you can ask questions about your code base without interfering with coding tasks.

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

## How It Works

YACT stores the current task in a plain text file: `.yact/prompt.txt` in your project directory. This file has two
sections separated by a `==========` line:

- Above the separator: a list of file paths that form the task context. The content of these files is sent to the LLM.
- Below the separator: your task prompt.

You build up a task by adding files with `y read`, writing your prompt into `.yact/prompt.txt`, then running one of
the mode commands (`ask`, `plan`, `act`). Each of these commands makes exactly one API call.

## Commands

### Start a New Task

Create a fresh, empty task context:

```bash
y new
```

### Add Files to the Context

Attach files so the AI can reference them:

```bash
y read main.go
y read "commands/*.go"
```

Glob patterns are supported. Directories are skipped, and files already in the context are not added twice.

### Ask Questions

Write your question into `.yact/prompt.txt`, then run:

```bash
y ask
```

The response is printed to the terminal. Your task context is not modified.

### Create a Plan

Write your task into `.yact/prompt.txt`, then run:

```bash
y plan
```

The generated plan is printed and also replaces the prompt in `.yact/prompt.txt`, so you can follow it up
directly with `y act`.

### Generate Code

Write your task (or generate a plan) into `.yact/prompt.txt`, then run:

```bash
y act
```

The AI responds with complete source files, which YACT writes directly to your filesystem. Newly created files are
automatically added to the task context. Use `--no-write` to print the response instead of writing files:

```bash
y -n act
```

### Edit a Line Range

Modify a specific line range in a file with a prompt, preserving indentation:

```bash
y snip main.go 10 20 "simplify this loop"
```

### Index the Project

Scan the project directory and create or update `.yact/index.csv`, which maps every file to a short AI-generated
description:

```bash
y index
```

Hidden files and common directories like `.git`, `node_modules`, `dist` and `build` are excluded. Deleted files
are removed from the index, and existing descriptions are preserved.

### Discover Relevant Files

Automatically populate the task context based on the prompt in `.yact/prompt.txt`:

```bash
y context
```

This uses the project index in two ways:

- Keyword matching: mark keywords in your prompt with double parentheses, e.g. `((authentication))`, and files whose
  index descriptions contain them are added.
- AI discovery: the file listing and your prompt are sent to the LLM, which selects the relevant files.

Run `y index` first, so the index is up to date.

## Global Flags

```
-h, --help        Show help message
-t, --think       Enable Claude's extended thinking mode
-n, --no-write    Do not write files, print response instead
-f, --fable       Use Claude Fable model
-o, --opus        Use Claude Opus model
-s, --sonnet      Use Claude Sonnet model
    --haiku       Use Claude Haiku model
```

Flags come before the command:

```bash
y -t act
y --think --sonnet plan
y -o ask
```

Model flags override the configured model for a single invocation. With `--think`, the model's reasoning is printed
before the response.

## Configuration

View current settings:

```bash
y config
```

Set configuration values:

```bash
y config anthropic_api_key your_key_here
y config claude_model sonnet
y config max_tokens 32000
y config think_budget 16000
```

Available configuration keys:
- `anthropic_api_key` - Your Claude API key (required)
- `claude_model` - Which Claude model to use: `fable`, `opus`, `sonnet` or `haiku` (default: `haiku`)
- `max_tokens` - Maximum output tokens per API call (default: 16000)
- `think_budget` - Token budget for extended thinking mode (default: 8000)

## System Prompts

Each mode uses a system prompt loaded from `~/.yact/systemprompts/`:

- `act.txt` - Code generation
- `plan.txt` - Planning
- `ask.txt` - Questions
- `snip.txt` - Line range editing
- `context.txt` - File discovery
- `index.txt` - File descriptions

These files must exist for the corresponding command to work, and you are free to tailor them to your workflow.

## Cost Tracking

Every API call prints the model used, call duration, input and output token counts, and the estimated cost in dollars.
A warning is shown if the response hit the output token limit and may be incomplete.

## Help

Display command reference:

```bash
y help
y --help
y -h
```

## Storage

Global settings are stored in `~/.yact/`:
- `config` - Your API key and model settings (JSON)
- `systemprompts/` - System prompt text files

Per-project state is stored in `.yact/` inside your project directory:
- `prompt.txt` - The current task context and prompt
- `index.csv` - File index with descriptions

## Troubleshooting

**"Claude API key not configured"**
- Set your API key: `y config anthropic_api_key YOUR_KEY`

**"No files found matching pattern"**
- Check that the glob pattern matches existing files
- Use exact paths if glob patterns don't work

**"no task prompt found"**
- Write your task below the `==========` separator in `.yact/prompt.txt` before running `y context`

**"Error loading ... prompt"**
- Create the missing system prompt file in `~/.yact/systemprompts/`

**API errors**
- Verify your API key is valid
- Check your internet connection
- Ensure your Claude API account has available credits