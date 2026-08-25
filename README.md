[![GitHub](https://img.shields.io/badge/GitHub-View%20Repository-blue?logo=github)](https://github.com/agabor/yact)

# YACT - Yet Another Coding Tool

Minimal, responsive, transparent LLM coding assistant.

## Design principles

- **Assisted, not autonomous.** YACT supports AI assisted software development, not AI driven software development. You decide what goes into the context and what the model is asked to do.
- **Responsive.** Each command is a single API call, so you get results in seconds rather than waiting on a multi-step agent loop.
- **Reactive.** YACT does what it is explicitly asked to do, and nothing more. It never explores your codebase or runs commands on its own initiative.
- **Transparent.** The full task context lives in a plain text file you can read and edit, and every response is logged verbatim.

## How YACT works

YACT is a CLI tool. You assemble a task by hand, send it once, and get one response back.

The current task lives in a plain text file, `.yact/prompt.txt`, in your project directory. It has two
sections separated by a `==========` line:

- **Above the separator:** a list of file paths that form the task context. The contents of these files are sent to the LLM.
- **Below the separator:** your task prompt.

You add files with `y read`, write your prompt, then run a prompt command such as `ask`, `plan`, `act` or
`bash`. Each of those makes exactly one API call. Responses are parsed for code blocks; complete source
files found in the response are written to disk, and any prose outside code blocks is printed to the terminal.

### Whole files, not diffs

YACT asks the model to emit complete source files rather than incremental edits. This keeps the LLM
interface simple — there are no editing tools to call, no patch formats to get wrong, and no partial-apply
failures — and it lets YACT keep each request self-contained and cheap.

The trade-off is that an editable file must fit inside the model's output token limit. Current Claude
models allow 64K–128K output tokens per response depending on the model, but YACT's own default
(`max_tokens`) is 16000, so raise it if you work with larger files. Keeping source files to a few hundred
lines is recommended regardless.

### Trade-offs to be aware of

YACT is a Swiss army knife for LLM assisted coding rather than a complete automated coding suite. It is
deliberately narrower than a full agent:

- You select the context yourself. YACT will not go looking for the files a task needs, so you need to
  know your codebase.
- No MCP support, no tool calling, no multi-turn agent loop.
- File sizes are bounded by the model's output token limit, as described above.
- Anthropic models only. There is currently no support for other providers or local models.

## Installation

Download the binary for your operating system from the
[GitHub releases page](https://github.com/agabor/yact/releases).

Available binaries:

- Linux: `yact-linux-amd64`
- Windows: `yact-windows-amd64.exe`
- macOS: `yact-darwin-amd64` or `yact-darwin-arm64`

The `releases/latest/download/` URLs below always resolve to the most recent release. If a download
fails, check the releases page for the current asset names.

On Linux:

```
cd /tmp
wget https://github.com/agabor/yact/releases/latest/download/yact-linux-amd64
chmod +x yact-linux-amd64
sudo mv yact-linux-amd64 /usr/local/bin/y
```

On macOS:

```
cd /tmp
curl -L -o yact-darwin-arm64 https://github.com/agabor/yact/releases/latest/download/yact-darwin-arm64
chmod +x yact-darwin-arm64
sudo mv yact-darwin-arm64 /usr/local/bin/y
```

Use `yact-darwin-amd64` instead on Intel Macs.

On Windows, download `yact-windows-amd64.exe` and place it in a directory on your PATH, or run it
directly from wherever you downloaded it.

### Building from source

YACT is written in Go. With a Go toolchain installed:

```
git clone https://github.com/agabor/yact.git
cd yact
go build -o y .
```

## Quick start

Configure your Anthropic API key:

```
y config anthropic_api_key YOUR_API_KEY
```

Verify your configuration at any time:

```
y config
```

Then start a task:

```
y new
y read main.go
y act "add a --version flag"
```

## Prompt commands

YACT is driven by system prompts. Every system prompt file in `~/.yact/systemprompts/` becomes a command.
For example, if `act.txt` exists, you can run `y act`. Typical prompt commands are `ask`, `plan`, `act`
and `bash`, but you can add your own. Each prompt command makes exactly one API call.

Download the official system prompts with the `-d` / `--download` flag:

```
y -d act
y -d plan
y -d ask
y -d bash
```

You can also write your own prompt files in `~/.yact/systemprompts/`, or tailor the official ones to your
workflow. Run `y help` to see the list of currently available prompt commands.

## Commands

### Start a new task

Create a fresh, empty task context:

```
y new
```

### Add files to the context

Attach files so the model can reference them:

```
y read main.go
y read "commands/*.go"
```

Glob patterns are supported. Directories are skipped, and files already in the context are not added twice.

### Add files by keyword

Recursively scan the project and add every file containing a keyword to the task context:

```
y keyword Transaction
```

Hidden directories are skipped, and only files with indexed extensions are considered (see
`y config ext` below). Files already in the context are not added twice.

### Set the prompt

Set the task prompt in `.yact/prompt.txt` directly from the CLI:

```
y prompt "implement the Foo feature"
```

With the `-b` / `--buffer` flag, the content of `.yact/buffer.txt` (the last LLM response) is used as the
prompt:

```
y -b prompt
```

### Run a prompt command

Write your task into `.yact/prompt.txt`, then run any prompt command:

```
y act
```

You can also pass the prompt inline as an optional argument, which replaces the prompt in
`.yact/prompt.txt` before the API call:

```
y act "implement Foo"
y ask "is Foo buggy?"
y plan "add authentication"
y bash "find all TODO comments in Go files"
```

The response is parsed for code blocks. Complete source files found in the response are written directly
to your filesystem, and newly created files are automatically added to the task context. Any text outside
code blocks is printed to the terminal. The raw response is always logged to `.yact/buffer.txt`.

Use `--no-write` to print the response instead of writing files:

```
y -n act
```

## Global flags

```
-h, --help        Show help message
-t, --think       Enable Claude's extended thinking mode
-n, --no-write    Do not write files, print response instead
-f, --fable       Use Claude Fable model
-o, --opus        Use Claude Opus model
-s, --sonnet      Use Claude Sonnet model
    --haiku       Use Claude Haiku model
-b, --buffer      Use the buffer content as the prompt
-d, --download    Download the system prompt for the command
```

Flags come before the command:

```
y -t act
y --think --sonnet plan
y -o ask
```

Model flags override the configured model for a single invocation. With `--think`, the model's reasoning
is printed before the response.

## Configuration

View current settings:

```
y config
```

Set configuration values:

```
y config anthropic_api_key your_key_here
y config claude_model sonnet
y config max_tokens 32000
y config think_budget 16000
```

Available configuration keys:

- `anthropic_api_key` — your Anthropic API key (required)
- `claude_model` — which Claude model to use: `fable`, `opus`, `sonnet` or `haiku` (default: `haiku`)
- `max_tokens` — maximum output tokens per API call (default: 16000)
- `think_budget` — token budget for extended thinking mode (default: 8000)

Note that `max_tokens` caps the size of what the model can write back. Since YACT generates complete
source files, this value bounds the size of the files it can edit. Raise it if responses are being
truncated; the ceiling is the model's own output limit.

### File extensions

The `y keyword` command only scans files with recognized extensions. A default set of common source file
extensions is stored in `.yact/extensions.txt`. Add an extension with:

```
y config ext vue
```

## Cost tracking

Every API call prints the model used, call duration, input and output token counts, and an estimated cost
in dollars. A warning is shown if the response hit the output token limit and may be incomplete.

## Storage

Global settings live in `~/.yact/`:

- `config` — API key and model settings (JSON)
- `systemprompts/` — system prompt text files

Per-project state lives in `.yact/` inside your project directory:

- `prompt.txt` — the current task context and prompt
- `buffer.txt` — log of the last LLM response
- `extensions.txt` — file extensions scanned by the keyword command

## Help

```
y help
y --help
y -h
```

## Troubleshooting

**"Claude API key not configured"**

- Set your API key: `y config anthropic_api_key YOUR_KEY`

**"No files found matching pattern"**

- Check that the glob pattern matches existing files
- Quote glob patterns so your shell does not expand them first
- Use exact paths if glob patterns don't work

**"Error loading ... prompt"**

- Download the missing system prompt: `y -d <command>`
- Or create the file manually in `~/.yact/systemprompts/`

**Truncated or incomplete responses**

- Raise `max_tokens`, or split the file you are editing into smaller ones

**API errors**

- Verify your API key is valid
- Check your internet connection
- Ensure your Anthropic API account has available credits
