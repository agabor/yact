# YACT — Yet Another Coding Tool

Minimal, responsive, transparent LLM coding assistant. Not for vibe coding.

Website and getting-started guide: [yact.dev](https://yact.dev) (sources in [`docs/`](docs/)).


> **Not affiliated with Anthropic.** YACT is an independent, open-source, third-party tool.
> It is not made, endorsed, or supported by Anthropic PBC. "Claude" and "Anthropic" are
> trademarks of Anthropic PBC; they are used here only to describe which API this tool
> calls. For Anthropic's own products and documentation, go to
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
- **Transparent.** The full task prompt lives in a plain Markdown file you can read and
  edit, and every response is logged verbatim.

## How YACT works

YACT is a CLI tool. You assemble a task by hand, send it once, and get one response back.

The task lives in a plain Markdown file, `yact-prompt.md`, in the root of your project
directory. It holds Markdown links to the files you want to send, and optionally a plan
or any other instructions you write yourself. The prompt file is sent to the LLM as the
prompt, and every project file whose path appears in the prompt is sent along as context.

You add files with `y read` (or `y keyword`), which append links such as
`[main.go](main.go)` to `yact-prompt.md`. Then you run a prompt command such as `ask`,
`plan`, `act` or `bash`, optionally passing an extra task prompt as an argument. Each of
those makes exactly one API call. Responses are parsed for code blocks; complete source
files found in the response are written to disk, and any prose outside code blocks is
printed to the terminal.

### Whole files, not diffs

YACT asks the model to emit complete source files rather than incremental edits. This keeps
the LLM interface simple — there are no editing tools to call, no patch formats to get
wrong, and no partial-apply failures — and it lets YACT keep each request self-contained
and cheap.

The trade-off is that an editable file must fit inside the model's output token limit.
YACT requests up to 128000 output tokens per call. See
[Anthropic's model documentation](https://docs.claude.com/en/docs/about-claude/models) for
the current per-model limits. Keeping source files to a few hundred lines is recommended
regardless.

### Deleting files

If a response contains a code block that has only a file path comment and no content, YACT
deletes that file from disk.

### What is sent

Every prompt command and `query` builds the request from these parts, in this order:

1. **Message files** in `.yact/`, if they exist: `domain.txt`, `index.txt` (see
   [Code index](#code-index)) and `output.txt` (see
   [Send command output](#send-command-output)).
2. **Context files**: every project file whose relative path appears in the prompt text.
3. **The prompt** itself.

The prompt is built from `yact-prompt.md` and the prompt argument:

- **With a prompt argument**, the content of `yact-prompt.md` (if any) is followed by the
  argument.
- **Without a prompt argument**, `yact-prompt.md` is the prompt. If it contains a plan with
  steps, only the steps that are not ticked off yet are sent (see
  [Plan, then act](#plan-then-act)). A missing or empty `yact-prompt.md` is an error.

### How context files are found

YACT scans the project and sends every file whose relative path, such as
`commands/call.go`, appears anywhere in the prompt text — whether as a Markdown link, in
backticks, or in plain prose.

- Hidden files and directories (names starting with `.`) are never sent, and neither is
  `yact-prompt.md` itself.
- Matching is a plain substring match on the path, so a mention of `src/router/router.go`
  also matches `router/router.go` if both exist.
- Each file is sent once, no matter how often it is mentioned.
- URLs are never fetched.

Use `--verbose` to print every path that is sent.

### Input size limit

Before every call YACT counts the lines of the prompt, the message files and all context
files and prints the total as `Input lines: <n>`. Only the text that is actually sent is
counted, so plan steps that are left out, whether through `--steps` or because they are
already done, are not counted. If the total exceeds `max_input_lines` (default: 2000), the
call is refused. Raise the limit permanently with `y config max_input_lines <n>`, or for a
single call with `-m <n>`.

### Trade-offs to be aware of

YACT is a Swiss army knife for LLM assisted coding rather than a complete automated coding
suite. It is deliberately narrower than a full agent:

- You select the context yourself. YACT will not go looking for the files a task needs, so
  you need to know your codebase.
- No MCP support, no tool calling, no multi-turn agent loop.
- File sizes are bounded by the model's output token limit, as described above.
- One provider only: Anthropic's API. There is currently no support for other providers or
  local models.

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

Download the official system prompts, then start a task:

```
y -d act
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

### Default flags in prompt files

A prompt file may start with a line of the form `flags: <letters>`. Each letter enables the
short form of a global flag whenever that prompt command runs. That first line is not sent
to the model. For example, a prompt file beginning with:

```
flags: c2n
```

behaves as if `-c 2 -n` had been passed. Supported letters:

- `n` — `--no-write` (ignored if `-v` is active)
- `c<level>` — Claude model level 1 to 4, for example `c1` for Haiku; a bare `c` or an
  invalid level selects Sonnet (ignored if you pass `-c` yourself)
- `i<level>` — `--index` level 1 or 2; a bare `i` selects 1 (ignored if you pass `-i`
  yourself)
- `d` — `--download`
- `v` — `--validate-code` (ignored if `-n` is active)
- `p` — `--plan`

Any other character is ignored. The official prompts use these defaults: `act` runs with
`c1v` (Haiku, validate code), `ask` with `c2n` (Sonnet, no write), `bash` with `c1n` (Haiku,
no write), and `plan` with `c2np` (Sonnet, no write, save the response as the plan).

## Commands

### Start a new task

Start over with an empty task:

```
y new
```

This deletes `yact-prompt.md` (including any plan in it) and the message files
`.yact/domain.txt`, `.yact/index.txt` and `.yact/output.txt`.

### Add files to the context

Attach files so the model can reference them:

```
y read main.go
y read "commands/*.go"
```

Glob patterns are supported. Each file is appended to `yact-prompt.md` as a Markdown link.
Directories are skipped, and files already linked are not added twice.

### Add files by keyword

Recursively scan the project and add every file containing a keyword to the task:

```
y keyword Transaction
```

Hidden directories are skipped, and only files with indexed extensions are considered (see
[File extensions](#file-extensions)). Files already linked are not added twice.

### Send command output

Read standard input into `.yact/output.txt`, which is sent with every call after the index
and before the context files:

```
go test ./... 2>&1 | y output
```

The file stays in place until it is overwritten by the next `y output` or removed by
`y new`.

### Code index

YACT can generate a compact index of the declarations in your project — file paths followed
by function, method, type and class signatures, without bodies. Supported languages are Go,
PHP, JavaScript, C#, Swift and Kotlin. `vendor`, `node_modules` and hidden directories are
skipped.

Generate an index for a single call with `-i` / `--index`:

```
y -i 1 ask "where is the config loaded?"
```

- `1` — public index: only the public interface is listed
- `2` — private index: all declarations are listed

The index is written to `.yact/index.txt` and sent only in the call that sets the flag.
Prompt commands and `query` without `-i` remove the index file before the call.

To write or remove the index without calling the LLM:

```
y index 1
y index 2
y index 0
```

### Run a prompt command

Run any prompt command with an optional task prompt argument:

```
y act "implement Foo"
y ask "is Foo buggy?"
y plan "add authentication"
y bash "find all TODO comments in Go files"
```

The prompt argument is appended to the content of `yact-prompt.md`, and the files
mentioned in either are sent as context (see [What is sent](#what-is-sent)). The argument
is not saved anywhere.

The response is parsed for code blocks. Complete source files found in the response are
written directly to your filesystem. Any text outside code blocks is printed to the
terminal. The raw response is always logged to `.yact/buffer.txt`.

Because responses are written to disk, run YACT inside a version-controlled working tree so
you can review and revert what it writes. Use `--no-write` to print the response instead of
writing files:

```
y -n act "implement Foo"
```

### Plan, then act

The `plan` prompt asks the model for a short Markdown plan. The plan has a `# Goal` section
and a `## Steps` section with one `### [ ] N. <title>` heading per step, and each step
changes exactly one file. Each step ends with a `**Context:**` line holding links to the
files needed for that step. With `-p` / `--plan`, the response is not processed as code; it
is saved to `yact-prompt.md`, replacing its previous content:

```
y -p plan "add authentication"
```

Then run a prompt command without a prompt argument:

```
y act
```

The plan is the prompt. Without `--steps`, the text before the first step (the goal) and
all steps that are not ticked off yet are sent; steps that are already done are left out,
and so are the files mentioned only in them. A plan without steps is sent as a whole. If
`yact-prompt.md` is missing or empty, the call fails.

Review or edit `yact-prompt.md` between the two steps if you want to adjust the plan.

#### Implementing selected steps

With `-s` / `--steps <selector>`, only the selected steps are implemented, one call at a
time. The goal and the selected steps are sent; all other steps are left out. The selector
can be:

- `N` — a single step, for example `2`
- `N-M` — a range of steps, for example `2-4`
- a comma separated list of the above, for example `1,3,5-6`
- `next` — the first step that is not ticked off yet

```
y -s next act
y -s 1,3 act
```

`--steps` requires a plan, and cannot be combined with a prompt argument or with `--plan`.
Plans with duplicate step numbers, selectors that name steps missing from the plan, and
plans without steps are rejected. Selecting a step that is already done prints a warning,
and the step is sent again.

#### Ticking off steps

After a call without a prompt argument has successfully written its response, the steps
that were sent are marked done in `yact-prompt.md`: `### [ ] 1. Title` becomes
`### [x] 1. Title`. When every step is already ticked, YACT prints
`All plan steps are done.` and makes no call. Steps are not ticked when `-n` is used, when
the call is submitted as a batch, or when the call fails. Untick a step by hand to run it
again.

#### Replacing the plan

Running a call with `-p` writes the response to `yact-prompt.md`, overwriting its previous
content, including any links added with `y read`. Nothing is written to your source tree
and the response is not validated in that case. Remove the plan with `y new` or by deleting
`yact-prompt.md`.

### Validate that a response is code only

With `-v` / `--validate-code`, the call fails if the response contains free text outside
code blocks or an incomplete code block. Nothing is written to disk in that case, although
the response is still logged to `.yact/buffer.txt`:

```
y -v act "implement Foo"
```

`-v` and `-n` are mutually exclusive when set through a prompt file's `flags:` line.

### Query without a system prompt

Call the LLM with no system prompt. The response is printed and never written to disk:

```
y query "explain what the keyword command does"
```

The prompt is built the same way as for prompt commands, and `query` honours `-p` and `-s`
the same way prompt commands do.

### Batch requests

With `-b` / `--batch`, the request is submitted to Anthropic's Message Batches API instead
of being streamed. Batches are billed at a 50% discount but may take a while to complete.
Only one batch can be pending at a time; its id is stored in `.yact/batch_job_id.txt`.

```
y -b act "implement Foo"
y batch
```

`y batch` prints the batch status while it is still processing. Once it has ended, the
result is fetched, token usage and the discounted cost are printed, the response is written
to `.yact/buffer.txt`, and the pending job is cleared. The output flags given at submission
are applied then:

- `-p` — the response is written to `yact-prompt.md`
- `-v` — the response is validated, and the call fails if it is not code only
- `-n` — the response is printed

Without these flags, the response is only stored in the buffer. Apply it with `y commit`.

### Commit the buffer

Write the code blocks of the last response in `.yact/buffer.txt` to disk, for example after
a `-n` run or a batch:

```
y -n act "implement Foo"
y commit
```

With `-v`, the commit fails if the buffer contains free text or an incomplete code block.

## Global flags

```
-h, --help                 Show help message
-n, --no-write             Do not write files, print response instead
-c, --claude <level>       Use Claude model (1: Haiku, 2: Sonnet, 3: Opus, 4: Fable)
-i, --index <level>        Index for this call: 0 sends no index, 1 generates a public
                           index, 2 generates a private index
-b, --batch                Submit the request as a batch job; the output flags (-n, -v,
                           -p) are applied when the result is fetched with 'y batch'
-d, --download             Download the system prompt for the command
    --no-progress          Hide progress indicator
-v, --validate-code        Fail if the response (or, for commit, the buffer) contains
                           free text or incomplete code blocks
-p, --plan                 Write the response to yact-prompt.md
-s, --steps <selector>     Implement only the selected steps of yact-prompt.md (N, N-M,
                           a comma separated list such as 1,3,5-6, or next). Requires a
                           plan; cannot be combined with a prompt argument or with --plan.
                           Finished steps are ticked off after a successful write
-m, --max-input-lines <n>  Set max input lines for the current command, without changing config
    --verbose              Print all file paths sent to the LLM
```

Flags come before the command:

```
y -c 3 ask "how does the parser work?"
y --claude 2 --no-progress plan "add authentication"
y -i 2 ask "where is the config loaded?"
y -m 3000 act "add a --version flag"
y -b act "add a --version flag"
y -p plan "add authentication"
y -s next act
```

Model flags override the configured model for a single invocation. Model availability
depends on your own Anthropic account. When a flag is used, YACT prints the effective flags
as `Executing: y <flags> <command>` before the call.

## Configuration

View current settings:

```
y config
```

Set configuration values:

```
y config claude_model sonnet
y config max_input_lines 3000
```

Available configuration keys:

- `anthropic_api_key` — your own Anthropic API key, stored locally (required)
- `claude_model` — which model to use: `fable`, `opus`, `sonnet` or `haiku`
  (default: `haiku`)
- `max_input_lines` — maximum number of lines of prompt, message files and context files
  sent in one call (default: 2000)

The maximum output token count per call is fixed at 128000.

### File extensions

The `y keyword` command only scans files with recognized extensions. A default set of
common source file extensions is stored in `.yact/extensions.txt`. Add an extension with:

```
y config ext vue
```

## Cost tracking

Every API call prints the model used, call duration, input and output token counts, and an
estimated cost in dollars. The estimate is calculated locally from published rates and is
informational only — your actual billing is whatever Anthropic charges your account. A
warning is shown if the response hit the output token limit and may be incomplete.

## Storage

Global settings live in `~/.yact/`:

- `config` — API key and model settings (JSON)
- `systemprompts/` — system prompt text files

Per-project state lives in your project directory:

- `yact-prompt.md` — the task prompt: links added by `y read` and `y keyword`, and the plan
  written by `-p`
- `.yact/buffer.txt` — log of the last LLM response
- `.yact/index.txt` — the code index, sent when present
- `.yact/output.txt` — command output captured by `y output`, sent when present
- `.yact/domain.txt` — optional domain description, sent when present
- `.yact/batch_job_id.txt` — the pending batch job and its output flags
- `.yact/extensions.txt` — file extensions scanned by the keyword command

Add `.yact/` to your `.gitignore` so task state and response logs are not committed.

`yact-prompt.md` is up to you. Ignore it if you treat it as a scratch file. Or keep it: a
plan that has been reviewed and ticked off can be attached to a pull request (committed,
pasted into the description, or kept elsewhere) to give reviewers more context. Each team
can decide how.

## Help

```
y help
y --help
y -h
```

## Troubleshooting

**"Claude API key not configured"**

- Set your key: `y config anthropic_api_key <your-own-anthropic-key>`

**"input exceeds maximum number allowed lines"**

- Start over with `y new` and re-add only the files you need
- Remove links you do not need from `yact-prompt.md`
- Select fewer plan steps with `-s`
- Raise the limit for one call with `-m <n>`, or permanently with
  `y config max_input_lines <n>`

**"response contains free text outside code blocks" / "incomplete code block"**

- These come from `-v` / `--validate-code`. Rerun the command, or drop `-v` to accept
  free text. The rejected response is still available in `.yact/buffer.txt`.

**"no plan found in yact-prompt.md, provide a prompt or create a plan with --plan"**

- You ran a call without a prompt argument and `yact-prompt.md` is missing or empty. Pass
  a prompt, add files with `y read`, or create a plan with `y -p plan "<task>"`.

**"--steps and --plan cannot be combined" / "--steps cannot be combined with a prompt argument"**

- `--steps` works on the saved plan only. Drop the prompt argument or the `-p` flag.

**"step range ... does not match the plan" / "invalid step selector item"**

- The selector names steps the plan does not contain, or is malformed. Use `N`, `N-M`, a
  comma separated list, or `next`, and check the numbers in `yact-prompt.md`.

**"a batch is already pending"**

- Fetch the pending result with `y batch` before submitting another batch.

**Unexpected files are sent**

- Every project file whose path appears in the prompt text is sent. Run with `--verbose`
  to see which files are sent, and rephrase or remove the mentions you do not want.

**"No files found matching pattern"**

- Check that the glob pattern matches existing files
- Quote glob patterns so your shell does not expand them first
- Use exact paths if glob patterns don't work

**"no system prompt found for command"**

- Download the missing system prompt: `y -d <command>`
- Or create the file manually in `~/.yact/systemprompts/`

**Truncated or incomplete responses**

- Split the file you are editing into smaller ones

**API errors**

- Verify your API key is valid in the Anthropic console
- Check your internet connection
- Ensure your account has available credits

## Security and privacy

YACT is a local command-line program. It has no server component and no website that
accepts input.

- **Your API key stays on your machine.** It is stored in `~/.yact/config` as plain JSON.
  YACT sends it only to `https://api.anthropic.com`, as the `x-api-key` header that
  Anthropic's API requires. It is never transmitted anywhere else.
- **You get your own key.** YACT never issues, requests, brokers, or collects API keys.
  Create an Anthropic key in your own account at
  [console.anthropic.com](https://console.anthropic.com).
- **No telemetry.** YACT does not phone home, collect analytics, or report usage. The only
  outbound network connections it makes are the API calls you explicitly trigger, and
  system-prompt downloads when you pass `-d`. URLs in the prompt are never fetched.
- **No background activity.** YACT runs when you invoke it and exits. It installs no
  service, daemon, or scheduled task, and modifies nothing outside `~/.yact/`, the
  `.yact/` directory and `yact-prompt.md` of the project you are working in, and the files
  written from responses.
- **You choose what is sent.** Only the files mentioned in your prompt (added with
  `y read` or `y keyword`, or written by you), the message files in `.yact/`, and the
  prompt itself are included in a request. YACT never scans or uploads your codebase on
  its own initiative; the code index is generated only when you pass `-i` or run
  `y index`.
- **Auditable.** The complete source is public. Every release is built from a tagged
  commit in the repository, and every response is logged verbatim to `.yact/buffer.txt`.

If you would rather not trust prebuilt binaries, build from source — it takes one command
and is documented under [Installation](#installation).

## Reporting a problem

Open an issue at
[github.com/agabor/yact/issues](https://github.com/agabor/yact/issues). For anything
security-sensitive, please report it privately through the repository's security advisory
page rather than in a public issue.

## Contributing

The codebase is small and split into a few focused packages (`api`, `commands`, `config`,
`index`, `logic`), so it is quick to read. Contributions are welcome: open an issue to
discuss an idea, or send a pull request.

You can also change what YACT does without touching the code: every `*.txt` file in
`~/.yact/systemprompts/` becomes a command. See [Prompt commands](#prompt-commands).

The website in `docs/` is plain HTML, CSS and JavaScript with no build step and no
external requests. `docs/assets/yact-scene.js` draws the hero animation.
