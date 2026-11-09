# YACT — Yet Another Coding Tool

Minimal, responsive, transparent LLM coding assistant. Not for vibe coding.

Website and getting-started guide: [yact.dev](https://yact.dev)

> **Not affiliated with Anthropic.** YACT is an independent, open-source, third-party tool.
> It is not made, endorsed, or supported by Anthropic PBC. "Claude" and "Anthropic" are
> trademarks of Anthropic PBC; they are used here only to describe which API this tool
> calls. For Anthropic's own products and documentation, go to
> [anthropic.com](https://www.anthropic.com) and
> [docs.claude.com](https://docs.claude.com).

## Contents

- [Design principles](#design-principles)
- [How YACT works](#how-yact-works)
- [Installation](#installation)
- [Quick start](#quick-start)
- [The prompt file](#the-prompt-file)
- [Prompt commands](#prompt-commands)
- [Plan, then act](#plan-then-act)
- [Commands](#commands)
- [Global flags](#global-flags)
- [Configuration](#configuration)
- [Storage](#storage)
- [Running YACT in GitHub Actions](#running-yact-in-github-actions)
- [Troubleshooting](#troubleshooting)
- [Security and privacy](#security-and-privacy)
- [Contributing](#contributing)

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

YACT is a CLI tool, invoked as `y`. You assemble a task by hand, send it once, and get one
response back.

The task lives in a plain Markdown file, `yact-prompt.md`, in the root of your project. It
holds your instructions (or a plan) and **links to the files you want to send**. When you
run a prompt command such as `ask`, `plan` or `act`, YACT:

1. builds one request from the prompt file, the linked files and a few optional message
   files,
2. sends it to Claude with the system prompt that belongs to the command,
3. saves the raw response to `.yact/temp/buffer.txt`,
4. and then, depending on the command, prints the response, saves it as a plan, or writes
   the files it contains to disk.

There is no tool calling, no MCP and no multi-turn loop.

### Whole files, not diffs

YACT asks the model to emit complete source files rather than incremental edits. This keeps
the LLM interface simple — no editing tools, no patch formats to get wrong, no partial-apply
failures — and keeps each request self-contained.

The trade-off is that an editable file must fit inside the model's output token limit.
YACT requests up to 128000 output tokens per call; see
[Anthropic's model documentation](https://docs.claude.com/en/docs/about-claude/models) for
per-model limits. YACT warns before the call about any linked file longer than 300 lines,
and the official prompts ask the model to keep files under about 200 lines.

A code block in a response is written to disk when its first line is a comment holding the
file path (`// path`, `# path`, `/* path */`, `-- path` or `<!-- path -->`, optionally after
a shebang line). A block with a path comment and **no content deletes that file**. Absolute
paths and paths that leave the project directory are rejected.

### What is sent

Every prompt command and `query` sends one user message made of these parts, in order:

1. `.yact/domain.txt` — an optional description of your project, written by you
2. `.yact/temp/index.txt` — the [code index](#code-index), when the call asks for one
3. `.yact/temp/output.txt` — command output captured with [`y output`](#send-command-output)
4. [Skills](#skills) attached with `--skill`
5. Every linked project file, as a code block whose first line is the file path
6. The prompt itself

Use `--verbose` to print every file and the prompt, and `--dry` to build and check the
request without calling the API.

### Trade-offs to be aware of

- You select the context yourself. YACT will not go looking for the files a task needs, so
  you need to know your codebase. The code index and keyword search help.
- File sizes are bounded by the model's output token limit.
- One provider only: Anthropic's API. There is no support for other providers or local
  models.

## Installation

### Build from source (recommended)

YACT is written in Go (1.26 or newer). Building from source means you run code you can
read:

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

There is also an `install.sh` script that builds the binary and moves it to
`/usr/local/bin` using `sudo`.

### Prebuilt binaries

Release binaries are built by GitHub Actions from tagged commits and published on the
[releases page](https://github.com/agabor/yact/releases):

| Platform | Asset |
| --- | --- |
| Linux (x86-64) | `yact-linux-amd64` |
| Linux (ARM64) | `yact-linux-arm64` |
| Windows (x86-64) | `yact-windows-amd64.exe` |
| macOS (Apple silicon) | `yact-darwin-arm64` |
| macOS (Intel) | `yact-darwin-amd64` |

On Linux and macOS:

```
chmod +x yact-linux-amd64
mkdir -p ~/.local/bin
mv yact-linux-amd64 ~/.local/bin/y
```

On Windows, place `yact-windows-amd64.exe` in a directory on your `PATH` (renaming it to
`y.exe` is convenient). Releases do not currently include checksum files; if you need to
verify what you run, build from source.

Check the installed version with `y --version`.

## Quick start

Create an API key in your own account at
[console.anthropic.com](https://console.anthropic.com) and store it. YACT reads the key
only from its own config file, not from the `ANTHROPIC_API_KEY` environment variable:

```
y config anthropic_api_key <your-own-anthropic-key>
y config
```

Then, inside a git working tree:

```
y new                                   # start an empty task
y read main.go                          # link a file into yact-prompt.md
y ask "how are flags parsed?"           # one call, answer printed
y act "add a --version flag"            # one call, files written to disk
git diff                                # review what was written
```

For anything bigger than a small change, plan first:

```
y new
y plan "add a --version flag"           # writes a plan to yact-prompt.md
$EDITOR yact-prompt.md                  # review or edit the plan
y act                                   # implement the plan, tick off its steps
```

System prompts are downloaded automatically from this repository the first time you use a
command.

## The prompt file

`yact-prompt.md` is both the prompt and the list of context files.

```
Add a --version flag that prints the version set at build time.
Keep the output format of ((help text)) unchanged.

-----

[main.go](main.go)
[[commands/*.go]]
```

### Selecting context files

Only explicit links select files. Plain mentions of a path in prose are not sent.

- **Markdown links** `[text](path)` send exactly that file. Wrap paths with spaces in angle
  brackets: `[my file.go](<my file.go>)`.
- **Double-bracket links** `[[pattern]]` send every project file matching a glob pattern,
  for example `[[commands/*.go]]`. Patterns follow Go's `filepath.Match`, so `*` does not
  cross `/`.
- **Keywords** `((phrase))` send every file whose content contains the phrase as whole
  words, ignoring case and separators: `((plan step))` matches `PlanStep`, `plan_step` and
  `plan step`, but not `explanation`. Only files with a configured
  [extension](#file-extensions) are searched. The parentheses are removed before the
  prompt is sent; the phrase itself stays.

Each file is sent once. Hidden files and folders are never sent, except the `.github`
folder so workflows can be edited. Folders listed in `.yact/ignore.txt` (by default
`vendor` and `node_modules`) are skipped, and so is `yact-prompt.md` itself. URLs are never
fetched.

### The footer

A line made only of dashes (such as `-----`) splits the file into a **body** and a
**footer**. Links in the footer select files, but the footer text is **not sent** to the
model. `y read` and `y keyword` add their links to the footer, creating it if needed, so
the instructions you write stay clean.

Note that any dash-only line counts, including a Markdown horizontal rule (`---`).

### How the prompt is built

- **With a prompt argument** (`y act "..."`), the argument is appended to the body of
  `yact-prompt.md` and the combined text is sent. The combined text is also **written back
  to `yact-prompt.md`**, so instructions accumulate across calls until you run `y new` or
  `y compact`.
- **Without a prompt argument** (`y act`), `yact-prompt.md` is the prompt. If it holds a
  plan, only some of its steps may be sent (see [Plan, then act](#plan-then-act)). A
  missing or empty file is an error.

## Prompt commands

Every system prompt file in `~/.yact/systemprompts/` becomes a command: `act.txt` gives you
`y act`. If a command's prompt is missing, YACT downloads it from the `systemprompts/`
folder of this repository's `main` branch. Pass `-d` / `--download` to download it again
and pick up the latest version. `y help` lists the prompts you have.

The official prompts:

| Command | Defaults | What it does |
| --- | --- | --- |
| `act` | `c1vi1` | Writes code. Responds with complete files only; fails on any free text |
| `plan` | `c1pi2` | Writes or updates a step-by-step plan in `yact-prompt.md` |
| `ask` | `c2n` | Answers questions about the code in plain text |
| `analyze` | `c2n` | Reports where the code breaks the code principles, with fixes |
| `review` | `c2n` | Reviews a diff sent with `y output` against the task |
| `bash` | `c1n` | Prints a plain list of shell commands for a task |

`act`, `plan`, `analyze` and `review` share one set of code principles
([`systemprompts/include/principles.txt`](systemprompts/include/principles.txt)), so code is
planned, written and reviewed against the same rules.

You can edit the official prompts or write your own. Prompt names may not contain `/`,
`\` or `..`.

### Default flags in prompt files

A prompt file may start with a line `flags: <letters>`. That line is not sent to the model.
Each letter turns on a flag whenever the command runs:

- `c<level>` — Claude model level 1 to 4 (`c1` is Haiku). A bare `c` or an invalid level
  selects 2 (Sonnet). Ignored when you pass `-c` yourself.
- `i<level>` — index level 1 or 2. A bare `i` selects 1. Ignored when you pass `-i`
  yourself, so `-i 0` turns the index off.
- `n` — `--no-write`
- `v` — `--validate-code`
- `p` — `--plan`
- `d` — `--download`

Other characters are ignored. The `n`, `v`, `p` and `d` defaults are always added to the
flags you pass, and `-n`, `-v` and `-p` exclude each other, so a combination such as
`y -n act` (act defaults to `v`) fails with an error. The effective flags are printed as
`Executing: y <flags> <command>` before the call.

### Includes

A prompt file can include shared text with `[[ name ]]`, which is replaced by the content
of `~/.yact/systemprompts/include/<name>`. Includes may include other files; cycles are
reported as errors. Includes are downloaded together with the prompt.

### Skills

`--skill <name>` attaches `~/.yact/skills/<name>.md` verbatim to the request, after the
message files and before the context files. It can be repeated. Missing skills are
downloaded from the `skills/` folder of this repository, and `-d` downloads them again.
The repository ships [`workflow`](skills/workflow.md), which teaches the model how to write
YACT GitHub workflows:

```
y --skill workflow act "add a workflow that runs y analyze on every pull request"
```

## Plan, then act

### Writing a plan

```
y plan "add authentication"
```

`plan` defaults to `-p`, so the response is saved to `yact-prompt.md`, replacing its
previous content (including links added with `y read`); nothing is written to your source
tree. The plan looks like this:

```
# Goal

One or two sentences.

## Steps

### [ ] 1. Create the User model

**File:** [src/models/user.go](src/models/user.go) (new)

- Add `type User struct { ... }`

**Context:** [src/db/connection.go](src/db/connection.go)

### [ ] 2. ...
```

Each step changes one file and links the files it needs, so the plan also selects the
context for each step. Review and edit it before you run it.

### Missing context

If the planner had to guess the content of a file it was not given, it lists that file
under `## Missing context`. A plan with that section **cannot be run**: `act` fails with
`Files needed as context are missing; run with --plan to update the plan`. Run `y plan`
again — the files linked in the section are now sent — and the planner revises the plan
and removes the section.

### Implementing the plan

```
y act
```

Without a prompt argument and without `--steps`, the goal and **every step that is not
ticked off yet** are sent; ticked steps, and files linked only in them, are left out. After
the files are written, the sent steps are ticked off in `yact-prompt.md`:
`### [ ] 1. Title` becomes `### [x] 1. Title`. Untick a step by hand to run it again.

Steps are ticked only when the response was written to disk — not with `-n`, not when the
call fails, and not for batch jobs. When every step is ticked, further calls fail with
`no steps remain; run with --plan to update the plan`.

### Implementing selected steps

With `-s` / `--steps <selector>`, only the selected steps are sent, together with the goal:

- `N` — a single step, for example `2`
- `N-M` — a range, for example `2-4`
- a comma separated list, for example `1,3,5-6`
- `next` — the first step that is not ticked off yet
- `0` — the whole plan as it is, without step selection; no step is ticked off

```
y -s next act
y -s 1,3 act
```

`--steps` cannot be combined with a prompt argument or with `--plan`. Selectors naming
steps the plan does not contain, and plans with duplicate step numbers, are rejected.
Selecting a step that is already done prints a warning and sends it again.

Links in the footer of a plan are part of its last step, so they are sent only when the
last step is.

### Updating a plan

Run `y plan` without an argument to update an existing plan: the planner checks which
steps are already implemented in the linked files, drops them and renumbers the rest. Add
an argument to change the plan:

```
y plan "also add a logout endpoint"
```

## Commands

```
y [flags] <command> [arguments]
```

| Command | Description |
| --- | --- |
| `<prompt command> [prompt]` | Call the LLM with the command's system prompt |
| `query [prompt]` | Call the LLM without a system prompt; the response is printed |
| `new` | Start a new task |
| `read <file>...` | Link files into `yact-prompt.md` (glob patterns supported) |
| `keyword <keyword>` | Link every file containing the keyword |
| `output` | Save standard input to be sent with every call |
| `index <level>` | Write or remove the code index without calling the LLM |
| `compact` | Reduce `yact-prompt.md` to its links |
| `commit` | Write the code blocks of the last response to disk |
| `batch` | Fetch the result of a pending batch job |
| `config [key value]` | Show or change configuration |
| `help` | Show help and the available prompt commands |

### Start a new task

```
y new
```

Deletes `yact-prompt.md` (including any plan in it), `.yact/temp/index.txt` and
`.yact/temp/output.txt`. `.yact/domain.txt` is kept.

### Link files

```
y read main.go
y read "commands/*.go"
y keyword Transaction
```

`y read` appends `[path](path)` links to the footer of `yact-prompt.md`. Quote glob
patterns so the shell does not expand them. Directories are skipped and files already
linked are not added twice.

`y keyword` links every project file whose content contains the keyword as whole
words (see [Selecting context files](#selecting-context-files)). Use `((keyword))` in the
prompt instead to search at call time without adding links.

### Send command output

```
go test ./... 2>&1 | y output
git diff | y output
```

Saves standard input to `.yact/temp/output.txt`. It is sent with **every** call until it
is overwritten by the next `y output` or removed by `y new`.

### Code index

YACT can generate a compact index of the declarations in your project: file paths followed
by function, method, type and class signatures, without bodies. Supported languages are
Go, PHP, JavaScript (`.js`, `.mjs`, `.cjs`), C#, Swift and Kotlin. Hidden and ignored
folders are skipped.

```
y -i 1 ask "where is the config loaded?"
```

- `0` — no index
- `1` — public index: only the public interface
- `2` — private index: all declarations

The index is written to `.yact/temp/index.txt` and sent only in the call that asks for it;
every other call deletes it. `act` defaults to level 1 and `plan` to level 2. To write or
remove the index without calling the LLM:

```
y index 2
y index 0
```

### Run a prompt command

```
y act "implement Foo"
y ask "is Foo buggy?"
y bash "find all TODO comments in Go files"
```

The response is always saved to `.yact/temp/buffer.txt`. What happens next depends on the
output mode:

- **default** — code blocks are written to disk, and any text outside them is printed
- `-v` / `--validate-code` — the same, but the call fails and nothing is written if the
  response contains free text, an incomplete code block or a rejected path
- `-n` / `--no-write` — the response is printed
- `-p` / `--plan` — the response is saved to `yact-prompt.md`

Run YACT inside a version-controlled working tree so you can review and revert what it
writes.

### Query without a system prompt

```
y query "explain what the keyword command does"
```

The prompt and context are built the same way as for prompt commands. The response is
always printed and never written, so `-p` and `-v` cannot be used with `query`. With no
argument, the plan in `yact-prompt.md` is the prompt and `-s` selects steps, but no step
is ticked off.

### Compact the prompt file

```
y compact
```

Replaces `yact-prompt.md` with just its links, once each, in a footer. The instructions are
dropped. Useful for reusing the context of a plan or a pull request description for a new
instruction.

### Commit the buffer

```
y commit
y -v commit
```

Writes the code blocks of the last response in `.yact/temp/buffer.txt` to disk and prints
any text around them. Use it after `-n`, or after a validated call that failed only because
the model added a sentence of prose. With `-v`, the commit fails if the buffer contains
free text or an incomplete code block.

### Batch requests

```
y -b act "implement Foo"
y batch
```

With `-b` / `--batch`, the request is submitted to Anthropic's Message Batches API instead
of being streamed. Batches cost half as much but may take a while. Only one batch can be
pending; it is stored in `.yact/batch_job_id.txt` together with the output mode.

`y batch` prints the status while the batch is processing. Once it has ended, it prints
token usage and the discounted cost, saves the response to the buffer, delivers it with the
output mode the batch was submitted with (writing files, validating, printing, or saving
the plan), and clears the pending job. Plan steps are not ticked off for batch jobs.

## Global flags

Flags come before the command.

```
-c, --claude <level>        Claude model: 1 Haiku, 2 Sonnet, 3 Opus, 4 Fable
-n, --no-write              Print the response instead of writing files
-v, --validate-code         Fail if the response (or, for commit, the buffer) contains
                            free text or incomplete code blocks
-p, --plan                  Write the response to yact-prompt.md
-s, --steps <selector>      Send only the selected plan steps: N, N-M, 1,3,5-6, next, or 0
-i, --index <level>         Code index for this call: 0 none, 1 public, 2 private
-b, --batch                 Submit the request as a batch job
    --skill <name>          Attach ~/.yact/skills/<name>.md; can be repeated
-d, --download              Download the command's system prompt and skills again
    --dry                   Build and check the request, skip the LLM call
    --verbose               Print the files and the prompt sent to the LLM
    --no-progress           Hide the progress indicator
    --version               Print the version and exit
-h, --help                  Show help
```

`-n`, `-v` and `-p` exclude each other.

```
y -c 3 ask "how does the parser work?"
y -i 2 ask "where is the config loaded?"
y -i 0 act "rename Foo to Bar"
y --dry --verbose act
y -b plan "add authentication"
```

## Configuration

```
y config                              # show settings
y config anthropic_api_key <key>      # required
y config claude_model sonnet          # default model
y config ext vue                      # add a file extension for keyword search
y config ignore dist                  # skip a root-level folder
```

Settings in `~/.yact/config`:

- `anthropic_api_key` — your own Anthropic API key
- `claude_model` — `haiku` (default), `sonnet`, `opus` or `fable`. The `-c` flag and a
  prompt's `c` default override it

The maximum output token count per call is fixed at 128000.

### File extensions

Keyword search (`y keyword` and `((keyword))`) only looks at files with the extensions in
`.yact/extensions.txt`, which is created with a default set of common source extensions the
first time it is needed. Add one with `y config ext <extension>`. An empty file means the
defaults.

### Ignored folders

`.yact/ignore.txt` lists root-level folders that are never linked, searched or indexed. It
is created with `vendor` and `node_modules`. Add one with `y config ignore <folder>`. An
empty file means no folder is ignored.

### Domain description

Write a short description of your project — what it does, its main concepts and conventions
— into `.yact/domain.txt`. It is sent first in every call.

### Cost tracking

Every call prints the model, duration, number of files sent, input and output tokens and an
estimated cost in dollars. The estimate is calculated locally from hard-coded rates and is
informational only. A warning is shown if the response hit the output token limit and may
be incomplete.

## Storage

Global, in `~/.yact/`:

- `config` — API key and default model (JSON)
- `systemprompts/` — system prompts, and `include/` for shared prompt parts
- `skills/` — skill files

Per project:

- `yact-prompt.md` — the task: instructions, links and plans
- `.yact/domain.txt` — optional project description, sent with every call
- `.yact/extensions.txt` — extensions searched by keyword
- `.yact/ignore.txt` — root-level folders to skip
- `.yact/batch_job_id.txt` — the pending batch job and its output mode
- `.yact/temp/buffer.txt` — the last LLM response
- `.yact/temp/index.txt` — the code index, when a call asks for one
- `.yact/temp/output.txt` — output captured by `y output`

A good `.gitignore` setup commits the `.yact/*.txt` settings so the whole team shares them,
and ignores the rest:

```
.yact/temp
.yact/batch_job_id.txt
yact-prompt.md
```

Whether you keep `yact-prompt.md` is up to you. A reviewed and ticked-off plan makes a good
pull request description.

## Running YACT in GitHub Actions

All state is in files and every call is one command, so YACT scripts easily. This
repository is developed with these workflows. Each one starts when an **action label** is
added to an issue or pull request:

| Workflow | Label | What it does |
| --- | --- | --- |
| [`yact-issue.yml`](.github/workflows/yact-issue.yml) | `implement` on an issue | Plans the issue, refines the plan if context is missing, implements it, commits with a generated message and opens a pull request with the plan as its description |
| [`yact-update.yml`](.github/workflows/yact-update.yml) | `update` on a PR | Updates the plan from the PR description and the latest comment, implements it and pushes |
| [`yact-fix.yml`](.github/workflows/yact-fix.yml) | `fix` on a PR | Feeds failing `go build`, `go test` and `go vet` output to `act` and pushes the fixes |
| [`yact-review.yml`](.github/workflows/yact-review.yml) | `review` on a PR | Sends the diff, the changed files and the linked issues to `review` and posts the result as a comment |

A second, optional label picks the model: `haiku`, `sonnet` or `opus`.

- Without a model label, `implement`, `update` and `fix` use Haiku and `review` uses
  Sonnet. If an issue or pull request has more than one model label, the strongest wins.
- The model label stays where you put it, so one `sonnet` label on a pull request applies
  to every later run on it.
- The workflow reads the labels when it starts, so add the model label first.
- The action label is removed when the run ends, whether it succeeded or not. Add it
  again to run the workflow again.

Create the labels once:

```
for label in implement update fix review haiku sonnet opus; do gh label create "$label"; done
```

They need an `ANTHROPIC_API_KEY` repository secret, and the runner needs access to
`raw.githubusercontent.com` to download the system prompts. Tips for CI:

- Pass `--no-progress`.
- Post `.yact/temp/buffer.txt` rather than scraping standard output.
- Debug a workflow without spending tokens with `--dry-run --verbose`.

To write your own workflow, use the [`workflow` skill](skills/workflow.md):
`y --skill workflow act "..."`.

## Troubleshooting

**"Claude API key not configured"**

- Set your key: `y config anthropic_api_key <your-own-anthropic-key>`.

**"--no-write and --validate-code cannot be combined"** (or `--plan` with either)

- The command's prompt file already turns one of these on (`act` uses `-v`, `plan` uses
  `-p`, `ask` uses `-n`). To see what `act` would write without writing it, use `--dry
  --verbose` to inspect the request, or rely on `git diff` afterwards.

**"response contains free text outside code blocks" / "incomplete code block"**

- These come from `-v`. Nothing was written. Rerun the command, or apply the response
  anyway with `y commit` after checking `.yact/temp/buffer.txt`.

**"no plan found in yact-prompt.md, provide a prompt or create a plan with --plan"**

- You ran a call without a prompt argument and `yact-prompt.md` is missing or empty. Pass a
  prompt, or create a plan with `y plan "<task>"`.

**"Files needed as context are missing; run with --plan to update the plan"**

- The plan has a `## Missing context` section. Run `y plan` once more.

**"no steps remain; run with --plan to update the plan"**

- Every step is ticked. Untick a step to rerun it, or update the plan with `y plan "..."`.

**"step range ... does not match the plan" / "invalid step selector item"**

- Use `N`, `N-M`, a comma separated list, `next` or `0`, and check the numbers in
  `yact-prompt.md`.

**"a batch is already pending"**

- Fetch the pending result with `y batch` first.

**Files are not sent, or unexpected files are sent**

- Only links select files. Run with `--dry --verbose` to see what would be sent. Check
  that the path is relative to the project root and not in a hidden or ignored folder.
  Remember that `.yact/temp/output.txt` is sent with every call until `y new`.

**The model keeps seeing old instructions**

- Prompt arguments are appended to `yact-prompt.md`. Edit the file, or run `y compact` or
  `y new`.

**"no system prompt found for command"**

- The prompt could not be downloaded. Check the command name and your connection, or create
  the file in `~/.yact/systemprompts/` yourself.

**"Warning: ... has N lines, more than 300" / truncated responses**

- Split large files: the model has to write them back in full.

**API errors**

- Check that your API key is valid in the Anthropic console, that your account has
  credits, and that the selected model is available to you.

## Security and privacy

YACT is a local command-line program. It has no server component.

- **Your API key stays on your machine.** It is stored in `~/.yact/config` as plain JSON
  and sent only to Anthropic's API, as the `x-api-key` header it requires.
- **You get your own key.** YACT never issues, requests, brokers or collects API keys.
- **No telemetry.** The only network connections YACT makes are the API calls you trigger,
  and downloads of system prompts and skills from `raw.githubusercontent.com` when they are
  missing locally or you pass `-d`. URLs in prompts are never fetched.
- **No background activity.** YACT runs when you invoke it and exits. It modifies nothing
  outside `~/.yact/`, the `.yact/` folder and `yact-prompt.md` of your project, and the
  files written from responses.
- **Writes stay in the project.** File paths in responses are parsed before anything is
  written or deleted; absolute paths and paths that leave the project directory are
  rejected.
- **You choose what is sent.** Only linked files, the message files in `.yact/`, attached
  skills and the prompt are sent. The code index is generated only when a call asks for it.
- **Auditable.** The source is public, releases are built from tagged commits, and every
  response is saved verbatim to `.yact/temp/buffer.txt`.

## Reporting a problem

Open an issue at [github.com/agabor/yact/issues](https://github.com/agabor/yact/issues).
For anything security-sensitive, please report it privately through the repository's
security advisory page rather than in a public issue.

## Contributing

The codebase is small and split into focused packages:

- `api` — the Claude client
- `commands` — one handler per command
- `config` — settings, flags and prompt loading
- `index` — the code indexer
- `logic` — prompt files, links, plans and response parsing

Run the tests with `go test ./...`. The code follows the same
[principles](systemprompts/include/principles.txt) that the prompts ask the model to
follow. Contributions are welcome: open an issue to discuss an idea, or send a pull
request.

You can also change what YACT does without touching the code: every `*.txt` file in
`~/.yact/systemprompts/` becomes a command.

The website in `docs/` is plain HTML, CSS and JavaScript with no build step.
`docs/assets/yact-scene.js` draws the hero animation.
