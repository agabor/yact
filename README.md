# YACT — Yet Another Coding Tool

Minimal, responsive, transparent LLM coding assistant.

**Source code:** [github.com/agabor/yact](https://github.com/agabor/yact)

**Author:** Gábor Angyal

**License:** GNU GPL v3, see [LICENSE](https://github.com/agabor/yact/blob/main/LICENSE) in the repository


> **Not affiliated with Anthropic or Amazon.** YACT is an independent, open-source,
> third-party tool. It is not made, endorsed, or supported by Anthropic PBC or Amazon Web
> Services. "Claude" and "Anthropic" are trademarks of Anthropic PBC, "AWS" and "Bedrock"
> are trademarks of Amazon.com, Inc.; they are used here only to describe which APIs this
> tool calls. For Anthropic's own products and documentation, go to
> [anthropic.com](https://www.anthropic.com) and
> [docs.claude.com](https://docs.claude.com).

## Design principles

- **Assisted, not autonomous.** YACT supports AI assisted software development, not AI
  driven software development. You decide what goes into the context and what the model is
  asked to do.
- **Responsive.** Each command is a single API call, so you get results in seconds rather
  than waiting on a multi-step agent loop.
- **Reactive.** YACT does what it is explicitly asked to do, and nothing more. It never
  explores your codebase or runs commands on its own initiative.
- **Transparent.** The full task context lives in a plain text file you can read and edit,
  and every response is logged verbatim.

## How YACT works

YACT is a CLI tool. You assemble a task by hand, send it once, and get one response back.

The current task lives in a plain text file, `.yact/prompt.txt`, in your project directory.
It has two sections separated by a `==========` line:

- **Above the separator:** a list of file paths that form the task context. The contents of
  these files are sent to the LLM.
- **Below the separator:** your task prompt.

You add files with `y read`, write your prompt, then run a prompt command such as `ask`,
`plan`, `act` or `bash`. Each of those makes exactly one API call. Responses are parsed for
code blocks; complete source files found in the response are written to disk, and any prose
outside code blocks is printed to the terminal.

### Whole files, not diffs

YACT asks the model to emit complete source files rather than incremental edits. This keeps
the LLM interface simple — there are no editing tools to call, no patch formats to get
wrong, and no partial-apply failures — and it lets YACT keep each request self-contained
and cheap.

The trade-off is that an editable file must fit inside the model's output token limit.
Current Claude models allow a large output budget per response, but YACT's own default
(`max_tokens`) is 16000, so raise it if you work with larger files. See
[Anthropic's model documentation](https://docs.claude.com/en/docs/about-claude/models) for
the current per-model limits. Keeping source files to a few hundred lines is recommended
regardless.

### Deleting files

If a response contains a code block that has only a file path comment and no content, YACT
deletes that file from disk and removes it from the task context.

### Trade-offs to be aware of

YACT is a Swiss army knife for LLM assisted coding rather than a complete automated coding
suite. It is deliberately narrower than a full agent:

- You select the context yourself. YACT will not go looking for the files a task needs, so
  you need to know your codebase.
- No MCP support, no tool calling, no multi-turn agent loop.
- File sizes are bounded by the model's output token limit, as described above.
- Two providers only: Anthropic's API and AWS Bedrock. There is currently no support for
  other providers or local models.

## Installation

### Build from source (recommended)

YACT is written in Go. Building from source means you run code you can read:

```
git clone https://github.com/agabor/yact.git
cd yact
go build -o y .
```

Then move `y` somewhere on your `PATH` — `~/.local/bin` needs no elevated privileges:

```
mkdir -p ~/.local/bin
mv y ~/.local/bin/
```

There is also an `install.sh` script that builds the binary and installs it to
`/usr/local/bin` using `sudo`.

### Prebuilt binaries

Release binaries are published as GitHub release artifacts, built from tagged commits in
the repository. Download them from the
[releases page](https://github.com/agabor/yact/releases):

| Platform | Asset |
| --- | --- |
| Linux (x86-64) | `yact-linux-amd64` |
| Windows (x86-64) | `yact-windows-amd64.exe` |
| macOS (Apple silicon) | `yact-darwin-arm64` |
| macOS (Intel) | `yact-darwin-amd64` |

Download the asset for your platform from the releases page, verify it against the checksum
published alongside it, make it executable, and place it on your `PATH`. On Linux and
macOS:

```
chmod +x yact-linux-amd64
mkdir -p ~/.local/bin
mv yact-linux-amd64 ~/.local/bin/y
```

Installing under `~/.local/bin` keeps the binary in your own user account. If you prefer a
system-wide install under `/usr/local/bin`, that requires administrator privileges — only
do so once you are satisfied with what you downloaded.

On Windows, place `yact-windows-amd64.exe` in a directory on your `PATH`, or run it
directly from wherever you downloaded it.

## Quick start

Configure your Anthropic API key. Create the key yourself at
[console.anthropic.com](https://console.anthropic.com); it is written to `~/.yact/config`
on your machine and used only for calls to Anthropic's API:

```
y config anthropic_api_key <your-own-anthropic-key>
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

YACT is driven by system prompts. Every system prompt file in `~/.yact/systemprompts/`
becomes a command. For example, if `act.txt` exists, you can run `y act`. Typical prompt
commands are `ask`, `plan`, `act` and `bash`, but you can add your own. Each prompt command
makes exactly one API call.

Download the official system prompts with the `-d` / `--download` flag. These are plain
text files fetched from the YACT repository; you can read them at any time in
`~/.yact/systemprompts/`:

```
y -d act
y -d plan
y -d ask
y -d bash
```

You can also write your own prompt files in `~/.yact/systemprompts/`, or tailor the
official ones to your workflow. Run `y help` to see the list of currently available prompt
commands.

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

Glob patterns are supported, and a tag name can be used in place of a path (see `y tag`
below). Directories are skipped, and files already in the context are not added twice.

### Add files by keyword

Recursively scan the project and add every file containing a keyword to the task context:

```
y keyword Transaction
```

Hidden directories are skipped, and only files with indexed extensions are considered (see
`y config ext` below). Files already in the context are not added twice.

### Tag files

Group files under a name so you can add them all at once later:

```
y tag api "api/*.go"
y read api
```

Tags are stored in `.yact/tags.csv`. The `read` command first checks whether its argument
matches a tag name, and falls back to glob matching if it does not.

### Set the prompt

Set the task prompt in `.yact/prompt.txt` directly from the CLI:

```
y prompt "implement the Foo feature"
```

With the `-b` / `--buffer` flag, the content of `.yact/buffer.txt` (the last LLM response)
is used as the prompt:

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

The response is parsed for code blocks. Complete source files found in the response are
written directly to your filesystem, and newly created files are automatically added to the
task context. Any text outside code blocks is printed to the terminal. The raw response is
always logged to `.yact/buffer.txt`.

Because responses are written to disk, run YACT inside a version-controlled working tree so
you can review and revert what it writes. Use `--no-write` to print the response instead of
writing files:

```
y -n act
```

## Global flags

```
-h, --help        Show help message
-t, --think       Enable extended thinking mode
-n, --no-write    Do not write files, print response instead
-f, --fable       Use the Claude Fable model
-o, --opus        Use the Claude Opus model
-s, --sonnet      Use the Claude Sonnet model
    --haiku       Use the Claude Haiku model
-w, --qmax        Use the Qwen3 235B A22B Instruct 2507 model on AWS Bedrock
-e, --qcoder      Use the Qwen3 Coder 30B A3B model on AWS Bedrock
-b, --buffer      Use the buffer content as the prompt
-d, --download    Download the system prompt for the command
-q, --quiet       Hide progress indicator
```

Flags come before the command:

```
y -t act
y --think --sonnet plan
y -o ask
```

Model flags override the configured model for a single invocation. With `--think`, the
model's reasoning is printed before the response; extended thinking is not supported by the
Bedrock models and is ignored there. Model availability depends on your own Anthropic or
AWS account.

## Configuration

View current settings:

```
y config
```

Set configuration values:

```
y config claude_model sonnet
y config max_tokens 32000
y config think_budget 16000
```

Available configuration keys:

- `anthropic_api_key` — your own Anthropic API key, stored locally (required for Claude
  models)
- `claude_model` — which model to use: `fable`, `opus`, `sonnet`, `haiku`, `qmax` or
  `qcoder` (default: `haiku`)
- `bedrock_model` — Bedrock model id used by `qmax`
  (default: `qwen.qwen3-235b-a22b-2507-v1:0`)
- `bedrock_coder_model` — Bedrock model id used by `qcoder`
  (default: `qwen.qwen3-coder-30b-a3b-v1:0`)
- `aws_region` — AWS region for Bedrock calls (default: `us-west-2`)
- `aws_api_key` — Bedrock API key, stored locally; when empty, the default AWS credential
  chain is used instead
- `max_tokens` — maximum output tokens per API call (default: 16000)
- `think_budget` — token budget for extended thinking mode (default: 8000)

Note that `max_tokens` caps the size of what the model can write back. Since YACT generates
complete source files, this value bounds the size of the files it can edit. Raise it if
responses are being truncated; the ceiling is the model's own output limit.

### AWS Bedrock

Selecting `qmax` or `qcoder`, either through configuration or the `-w` / `-e` flags, routes
the call to AWS Bedrock instead of Anthropic. Authentication works in one of two ways:

- Set `aws_api_key` and YACT sends it as a bearer token, bypassing SigV4 signing.
- Leave `aws_api_key` empty and YACT uses the standard AWS credential chain (environment
  variables, shared config, instance roles).

The region comes from `aws_region`, and the model id from `bedrock_model` or
`bedrock_coder_model` depending on which flag you used.

### File extensions

The `y keyword` command only scans files with recognized extensions. A default set of
common source file extensions is stored in `.yact/extensions.txt`. Add an extension with:

```
y config ext vue
```

## Cost tracking

Every API call prints the model used, call duration, input and output token counts, and an
estimated cost in dollars. The estimate is calculated locally from published rates and is
informational only — your actual billing is whatever Anthropic or AWS charges your account.
A warning is shown if the response hit the output token limit and may be incomplete.

## Storage

Global settings live in `~/.yact/`:

- `config` — API keys and model settings (JSON)
- `systemprompts/` — system prompt text files

Per-project state lives in `.yact/` inside your project directory:

- `prompt.txt` — the current task context and prompt
- `buffer.txt` — log of the last LLM response
- `tags.csv` — file tags created with `y tag`
- `extensions.txt` — file extensions scanned by the keyword command

Add `.yact/` to your `.gitignore` so task context and response logs are not committed.

## Help

```
y help
y --help
y -h
```

## Troubleshooting

**"Claude API key not configured"**

- Set your key: `y config anthropic_api_key <your-own-anthropic-key>`

**"unable to load AWS configuration"**

- Configure AWS credentials and region, or set a Bedrock key:
  `y config aws_api_key <your-own-key>`
- Check that `aws_region` matches a region where the model is available

**"No files found matching pattern"**

- Check that the glob pattern matches existing files
- Quote glob patterns so your shell does not expand them first
- Use exact paths if glob patterns don't work

**"no system prompt found for command"**

- Download the missing system prompt: `y -d <command>`
- Or create the file manually in `~/.yact/systemprompts/`

**Truncated or incomplete responses**

- Raise `max_tokens`, or split the file you are editing into smaller ones

**API errors**

- Verify your API key is valid in the Anthropic or AWS console
- Check your internet connection
- Ensure your account has available credits

## Security and privacy

YACT is a local command-line program. It has no server component and no website that
accepts input.

- **Your API keys stay on your machine.** They are stored in `~/.yact/config` as plain
  JSON. YACT sends the Anthropic key only to `https://api.anthropic.com`, as the
  `x-api-key` header that Anthropic's API requires, and the AWS key only to the Bedrock
  endpoint for your configured region. They are never transmitted anywhere else.
- **You get your own keys.** YACT never issues, requests, brokers, or collects API keys.
  Create an Anthropic key in your own account at
  [console.anthropic.com](https://console.anthropic.com), and manage AWS credentials in
  your own AWS account.
- **No telemetry.** YACT does not phone home, collect analytics, or report usage. The only
  outbound network connections it makes are the API calls you explicitly trigger, and
  system-prompt downloads when you pass `-d`.
- **No background activity.** YACT runs when you invoke it and exits. It installs no
  service, daemon, or scheduled task, and modifies nothing outside `~/.yact/` and the
  `.yact/` directory of the project you are working in.
- **You choose what is sent.** Only the files you add with `y read`, `y tag` or
  `y keyword`, plus the prompt you write, are included in a request. YACT never scans or
  uploads your codebase on its own initiative.
- **Auditable.** The complete source is public. Every release is built from a tagged
  commit in the repository, and every response is logged verbatim to `.yact/buffer.txt`.

If you would rather not trust prebuilt binaries, build from source — it takes one command
and is documented under [Installation](#installation).

## Reporting a problem

Open an issue at
[github.com/agabor/yact/issues](https://github.com/agabor/yact/issues). For anything
security-sensitive, please report it privately through the repository's security advisory
page rather than in a public issue.
