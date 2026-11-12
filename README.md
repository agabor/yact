# YACT — Yet Another Coding Tool

**Lint the prompt. Review the change against its plan.**

YACT is a command-line tool for AI-assisted coding that works one way: plan first. You write
a prompt. The model turns it into a plan and lists what your prompt left unclear. You
resolve every entry, and one model call then implements each step. The plan goes to the
reviewer together with the change. Not for vibe coding.

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
- [No automated workflows](#no-automated-workflows)
- [Troubleshooting](#troubleshooting)
- [Security and privacy](#security-and-privacy)
- [Contributing](#contributing)

## Design principles

- **A prompt is a program.** A prompt is a program in a very high-level language, so it can
  be checked before it runs. When you plan, the model checks your prompt for two properties
  and writes the result into the plan. The prompt must be **complete**: code the model
  needs and was not given is listed under `## Missing context`. It must have **no silent
  assumptions**: every decision the model made for you is listed under `## Assumptions`.
  Both block `act` until you resolve them. See [Plan, then act](#plan-then-act).
- **A person is present for every decision.** Only you accept an assumption. YACT has no
  agent loop and [no automated workflows](#no-automated-workflows).
- **The plan goes with the change.** Each step names the file it changes and links the
  context it was given. A reviewer who has the plan compares the change with what was asked.
- **Predictable.** Each command is a single API call. YACT never explores your codebase,
  runs commands or retries on its own initiative.
- **Transparent.** The prompt and the plan live in a plain Markdown file you can read and
  edit, every response is saved verbatim, and the cost of every call is printed.

The lint is not a proof. The model does the checking and can overlook something, so you
still read the plan, check the change and run your tests. More on the method:
[The YACT way](https://yact.dev/the-yact-way.html).

## How YACT works

YACT is a CLI tool, invoked as `y`. You write a prompt and link the files it needs, send it
once, and get one response back.

The prompt lives in a plain Markdown file, `yact-prompt.md`, in the root of your project. It
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

Use `--verbose` to print every file and the prompt, and `--dry-run` to build and check the
request without calling the API.

### Trade-offs to be aware of

- You select the context yourself. YACT will not go looking for the files a change needs,
  so you need to know your codebase. The code index and keyword search help, and the plan
  lists what is still missing.
- You resolve the plan. `act` does not run until the missing context is supplied and every
  assumption is accepted or answered.
- File sizes are bounded by the model's output token limit, and every changed file is
  written back in full. Keep files small.
- YACT does not run your tests and does not run unattended.
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

Then, inside a git working tree, plan first:

```
y new                                   # start an empty prompt
y read main.go                          # link a file into yact-prompt.md
y plan "add a --version flag"           # one call, writes a plan to yact-prompt.md
$EDITOR yact-prompt.md                  # read the plan, resolve its two sections
y -s next act                           # one call, implements the next step
git diff                                # check the change against the step
```

Repeat the last two commands until every step is ticked. For a trivial change you can skip
the plan, and to understand code before you change it, ask:

```
y act "rename Foo to Bar"               # one call, files written to disk
y ask "how are flags parsed?"           # one call, answer printed
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
folder. Folders listed in `.yact/ignore.txt` (by default
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
folder of this repository's `main` branch. `y help` lists the prompts you have.

To replace your copy with the latest version, download it again. This overwrites any edits
you made to the file:

```
y download prompt act
```

The official prompts:

| Command | Defaults | What it does |
| --- | --- | --- |
| `plan` | `--model haiku --output plan --index all` | Writes or updates a step-by-step plan in `yact-prompt.md`, with the lint of the prompt |
| `act` | `--model haiku --strict --index public` | Writes code. Responds with complete files only; fails on any free text |
| `ask` | `--model sonnet --output print` | Answers questions about the code in plain text |
| `analyze` | `--model sonnet --output print` | Reports where the code breaks the code principles, with fixes |
| `review` | `--model sonnet --output print` | Reviews a diff sent with `y output` against the prompt |
| `bash` | `--model haiku --output print` | Prints a plain list of shell commands |

`act`, `plan`, `analyze` and `review` share one set of code principles
([`systemprompts/include/principles.txt`](systemprompts/include/principles.txt)), so code is
planned, written and reviewed against the same rules. They are described on
[yact.dev](https://yact.dev/code-principles.html).

You can write your own prompts. Before you edit an official one, copy it to a new name,
because `y download prompt` overwrites the original. Prompt names may not contain `/`, `\`
or `..`, and the names of the built-in [commands](#commands) are taken.

### Default flags in prompt files

A prompt file may start with a line `flags: ` followed by flags. That line is not sent to
the model. It sets the defaults of the command:

```
flags: --model haiku --strict --index public
```

Only `--model`, `--index`, `--output` and `--strict` are allowed here. Flags you pass on the
command line win over these defaults: `y -o print act` prints the response even though
`act` writes files by default. A `--strict` default applies only when the output is
`apply`. The effective flags are printed as `Executing: y <flags> <command>` before the
call.

An older letter syntax is still read, so existing prompt files keep working: `c1` to `c4`
for the model, `i1` or `i2` for the index, `n` for `--output print`, `p` for `--output plan`
and `v` for `--strict`, as in `flags: c1vi1`.

### Includes

A prompt file can include shared text with `[[ name ]]`, which is replaced by the content
of `~/.yact/systemprompts/include/<name>`. Includes may include other files; cycles are
reported as errors. Includes are downloaded together with the prompt.

### Skills

`--skill <name>` attaches `~/.yact/skills/<name>.md` verbatim to the request, after the
message files and before the context files. It can be repeated or comma separated. A skill
is a file you write: reusable instructions that only some calls need, such as a house style.
A skill that is missing locally is looked up in the `skills/` folder of this repository,
and `y download skill <name>` downloads it again.

## Plan, then act

### Writing a plan

```
y plan "add authentication"
```

`plan` defaults to `--output plan`, so the response is saved to `yact-prompt.md`, replacing
its previous content (including links added with `y read`); nothing is written to your
source tree. The plan looks like this:

```
# Goal

One or two sentences.

## Steps

### [ ] 1. Create the User model

**File:** [src/models/user.go](src/models/user.go) (new)

- Add `type User struct { ... }`

**Context:** [src/db/connection.go](src/db/connection.go)

### [ ] 2. ...

## Assumptions

- [ ] **A1. Unclear:** the prompt does not say how long a session lasts. Taken: 24 hours.

## Missing context

- [src/config/config.go](src/config/config.go): the name of the session secret setting
```

Each step changes one file and links the files it needs, so the plan also selects the
context for each step. The two sections at the end are the lint of your prompt. Each is
left out when it has no entries, and each **blocks `act`** until you resolve it.

### Missing context

If the model had to guess the content of a file it was not given, it lists that file under
`## Missing context`. A plan with that section cannot be run: `act` fails with
`Files needed as context are missing; run with --plan to update the plan`. Run `y plan`
again. The files linked in the section are sent this time, and the model revises the plan
and removes the section. You can also link files yourself with `y read`, `y keyword` or a
`((keyword))` in the prompt.

### Assumptions

Where the prompt leaves a decision open, the model makes it and lists it under
`## Assumptions`. Each entry has a checkbox and an id, and you resolve every entry before
`act` runs:

```
## Assumptions

- [x] **A1. Unclear:** the prompt does not say how long a session lasts. Taken: 24 hours.
- [ ] **A2. Unclear:** the prompt does not say where errors are shown. Taken: on a new page.
  Answer: inline, after the field loses focus
```

- **Accept** an entry by ticking its box by hand: change `[ ]` to `[x]`.
- **Answer** an entry by writing a line that starts with `Answer:` under it, then run
  `y plan` again. The model changes the steps to follow your answer and rewrites the entry
  with `Decided:` in place of `Taken:`. Its box is still open: read how the model understood
  you, then tick it.

While an entry is open and has no answer, `act` fails with `Assumptions are still open`.
While an entry has an answer that has not been planned yet, it fails with
`Answers under assumptions are not applied yet`. `y plan` is never blocked.

Only you tick a box. The model always writes open boxes, and after each plan call YACT
decides every box: a box is ticked only if the entry with the same id was ticked in the
previous plan and its text is unchanged. If the model rewords an entry you accepted, you
accept it again. A list item without a box counts as open.

The two sections hold different things. Missing context holds facts about code the model
could not see, and you resolve it by linking files. Assumptions holds decisions about what
you want, and you resolve it by deciding.

### Implementing the plan

```
y -s next act
```

This sends the goal and the first step that is not ticked off yet. After the files are
written, the step is ticked off in `yact-prompt.md`: `### [ ] 1. Title` becomes
`### [x] 1. Title`. Check the change with `git diff`, then run the command again for the
next step. Untick a step by hand to run it again.

Without `--steps`, `y act` sends the goal and **every step that is not ticked off yet** in
one call; ticked steps, and files linked only in them, are left out.

Steps are ticked only when the response was written to disk: not with `--output print`,
not when the call fails, and not for batch jobs. When every step is ticked, further calls
fail with `no steps remain; run with --plan to update the plan`.

If `act` finds a problem in the plan while it writes the code, such as a contradiction or a
choice the plan left open, it writes a note to `.yact/temp/feedback.md`, and YACT prints it
after the call. Treat the note as a bug report on the prompt. The file is not cleared
between calls.

### Implementing selected steps

With `-s` / `--steps <selector>`, only the selected steps are sent, together with the goal:

- `N` — a single step, for example `2`
- `N-M` — a range, for example `2-4`
- a comma separated list, for example `1,3,5-6`
- `next` — the first step that is not ticked off yet
- `all` (or `0`) — the whole plan as it is, without step selection; no step is ticked off

```
y -s next act
y -s 1,3 act
```

`--steps` cannot be combined with a prompt argument or with `--output plan`. Selectors
naming steps the plan does not contain, and plans with duplicate step numbers, are
rejected. Selecting a step that is already done prints a warning and sends it again. The
two sections block `act` whichever steps you select.

Links in the footer of a plan are part of its last step, so they are sent only when the
last step is.

### Updating a plan

Run `y plan` without an argument to update an existing plan: the model checks which steps
are already implemented in the linked files, drops them and renumbers the rest. It also
applies your answers to assumptions. Add an argument to change the plan:

```
y plan "also add a logout endpoint"
```

### Keeping the plan

YACT does not attach the plan to anything. It is in `yact-prompt.md`, and `y new` deletes
that file. Commit the plan with the change or copy it into the pull request, whichever your
team prefers, so that the reviewer can compare the change with it.

## Commands

```
y [flags] <command> [arguments]
```

| Command | Description |
| --- | --- |
| `<prompt command> [prompt]` | Call the LLM with the command's system prompt |
| `query [prompt]` | Call the LLM without a system prompt; the response is printed |
| `new` | Start a new, empty prompt |
| `read <file>...` | Link files into `yact-prompt.md` (glob patterns supported) |
| `keyword <keyword>` | Link every file containing the keyword |
| `output` | Save standard input to be sent with every call |
| `index <none\|public\|all>` | Write or remove the code index without calling the LLM |
| `compact` | Reduce `yact-prompt.md` to its links |
| `commit` | Write the code blocks of the last response to disk |
| `batch` | Fetch the result of a pending batch job |
| `download prompt <command>` | Download a system prompt again, replacing the local copy |
| `download skill <name>` | Download a skill again, replacing the local copy |
| `config [key value]` | Show or change configuration |
| `help` | Show help and the available prompt commands |

### Start a new prompt

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
y -i public ask "where is the config loaded?"
```

- `none` — no index
- `public` — only the public interface
- `all` — all declarations

The numbers `0`, `1` and `2` are accepted too. The index is written to
`.yact/temp/index.txt` and sent only in the call that asks for it; every other call deletes
it. `act` defaults to `public` and `plan` to `all`. With the index, the model can name code
it was not given and ask for it under `## Missing context`. To write or remove the index
without calling the LLM:

```
y index all
y index none
```

### Run a prompt command

```
y act "implement Foo"
y ask "is Foo buggy?"
y bash "find all TODO comments in Go files"
```

The response is always saved to `.yact/temp/buffer.txt`. What happens next depends on the
output mode, set with `-o` / `--output` or by the command's prompt file:

- `apply` (the default) — code blocks are written to disk, and any text outside them is
  printed
- `apply` with `--strict` — the same, but the call fails and nothing is written if the
  response contains free text, an incomplete code block or a rejected path
- `print` — the response is printed
- `plan` — the response is saved to `yact-prompt.md`

Run YACT inside a version-controlled working tree so you can review and revert what it
writes.

### Query without a system prompt

```
y query "explain what the keyword command does"
```

The prompt and context are built the same way as for prompt commands. The response is
printed by default. With no argument, the plan in `yact-prompt.md` is the prompt and `-s`
selects steps, but no step is ticked off when the response is only printed.

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
y --strict commit
```

Writes the code blocks of the last response in `.yact/temp/buffer.txt` to disk and prints
any text around them. Use it after `--output print`, or after a strict call that failed
only because the model added a sentence of prose. With `--strict`, the commit fails if the
buffer contains free text or an incomplete code block.

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
output mode the batch was submitted with (writing files, strict or not, printing, or saving
the plan), and clears the pending job. Plan steps are not ticked off for batch jobs.

## Global flags

Flags come before the command.

```
-m, --model <name>          Model: haiku, sonnet, opus, fable, or 1 to 4
-o, --output <mode>         apply writes files, print prints the response, plan writes it
                            to yact-prompt.md
    --strict                Fail if the response (or, for commit, the buffer) contains
                            free text or incomplete code blocks. Only with apply output
-s, --steps <selector>      Send only the selected plan steps: N, N-M, 1,3,5-6, next or all
-i, --index <name>          Code index for this call: none, public or all
-b, --batch                 Submit the request as a batch job
    --skill <name>          Attach ~/.yact/skills/<name>.md; repeatable or comma separated
    --dry-run               Build and check the request, skip the LLM call
    --verbose               Print the files and the prompt sent to the LLM
    --no-progress           Hide the progress indicator
-V, --version               Print the version and exit
-h, --help                  Show help
```

```
y -m opus plan "add authentication"
y -i all ask "where is the config loaded?"
y -i none act "rename Foo to Bar"
y -o print act
y --dry-run --verbose act
y -b plan "add authentication"
```

## Configuration

```
y config                              # show settings
y config anthropic_api_key <key>      # required
y config claude_model sonnet          # default model
y config ext vue                      # add a file extension for keyword search
y config ignore dist                  # skip a folder
```

Settings in `~/.yact/config`:

- `anthropic_api_key` — your own Anthropic API key
- `claude_model` — `haiku` (default), `sonnet`, `opus` or `fable`. The `-m` flag and the
  `--model` default of a prompt file override it

The maximum output token count per call is fixed at 128000.

### File extensions

Keyword search (`y keyword` and `((keyword))`) only looks at files with the extensions in
`.yact/extensions.txt`, which is created with a default set of common source extensions the
first time it is needed. Add one with `y config ext <extension>`. An empty file means the
defaults.

### Ignored folders

`.yact/ignore.txt` lists folders that are never linked, searched or indexed. Each line is a
folder path relative to the project root, with forward slashes. A path segment may use `*`
wildcards, and `**` as a whole segment matches any number of folders, as in
`build/**/generated`. The file is created with `vendor` and `node_modules`. Add a line with
`y config ignore <folder>`. An empty file means no folder is ignored.

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

- `yact-prompt.md` — the prompt: your text and links, or the plan made from them
- `.yact/domain.txt` — optional project description, sent with every call
- `.yact/extensions.txt` — extensions searched by keyword
- `.yact/ignore.txt` — folders to skip
- `.yact/batch_job_id.txt` — the pending batch job and its output mode
- `.yact/temp/buffer.txt` — the last LLM response
- `.yact/temp/index.txt` — the code index, when a call asks for one
- `.yact/temp/output.txt` — output captured by `y output`
- `.yact/temp/feedback.md` — the note `act` writes when it finds a problem in the plan

A good `.gitignore` setup commits the `.yact/*.txt` settings so the whole team shares them,
and ignores the rest:

```
.yact/temp
.yact/batch_job_id.txt
yact-prompt.md
```

Whether you commit `yact-prompt.md` is up to you. The plan should reach the reviewer one
way or the other: commit it with the change and leave it out of `.gitignore`, or copy it
into the pull request.

## No automated workflows

YACT does not support running unattended, and the GitHub workflows that used to be in this
repository are removed. A plan needs a person to accept or answer its assumptions, and a
job that plans and implements with nobody reading would have to tick the boxes itself.

You can still call single commands from your own scripts. `--no-progress` keeps logs
readable, the response of every call is in `.yact/temp/buffer.txt`, and errors go to
standard error with exit code 1.

## Troubleshooting

**"Claude API key not configured"**

- Set your key: `y config anthropic_api_key <your-own-anthropic-key>`.

**"--strict cannot be combined with --output print"** (or `plan`)

- `--strict` only makes sense when files are written. To read what `act` would produce
  without writing it, use `y -o print act`. To see what it would send, use
  `--dry-run --verbose`.

**"response contains free text outside code blocks" / "incomplete code block"**

- These come from `--strict`, which `act` turns on by default. Nothing was written. Rerun
  the command, or apply the response anyway with `y commit` after checking
  `.yact/temp/buffer.txt`.

**"no plan found in yact-prompt.md, provide a prompt or create a plan with --plan"**

- You ran a call without a prompt argument and `yact-prompt.md` is missing or empty. Pass a
  prompt, or create a plan with `y plan "<prompt>"`.

**"Files needed as context are missing; run with --plan to update the plan"**

- The plan has a `## Missing context` section. Run `y plan` once more, or link the files
  yourself.

**"Assumptions are still open"**

- An entry under `## Assumptions` has an open box and no answer. Tick the box if you accept
  the entry, or write an `Answer:` line under it and run `y plan`.

**"Answers under assumptions are not applied yet"**

- You answered an assumption and have not planned again. Run `y plan`, read the rewritten
  entry and tick it.

**An assumption you ticked is open again**

- The model reworded the entry, so the tick was not kept. Read the new wording and tick it
  again.

**"no steps remain; run with --plan to update the plan"**

- Every step is ticked. Untick a step to rerun it, or update the plan with `y plan "..."`.

**"step range ... does not match the plan" / "invalid step selector item"**

- Use `N`, `N-M`, a comma separated list, `next` or `all`, and check the numbers in
  `yact-prompt.md`.

**"a batch is already pending"**

- Fetch the pending result with `y batch` first.

**Files are not sent, or unexpected files are sent**

- Only links select files. Run with `--dry-run --verbose` to see what would be sent. Check
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
  missing locally or you run `y download`. URLs in prompts are never fetched.
- **No background activity.** YACT runs when you invoke it and exits. It modifies nothing
  outside `~/.yact/`, the `.yact/` folder and `yact-prompt.md` of your project, and the
  files written from responses.
- **Writes stay in the project.** File paths in responses are parsed before anything is
  written or deleted; absolute paths and paths that leave the project directory are
  rejected. Inside the project, a response can write or delete any file, not only the
  linked ones, so check `git diff` before you commit.
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
- `logic` — prompt files, links, plans, assumptions and response parsing
- `projectpath` — the type for paths inside the project

Run the tests with `go test ./...`. The code follows the same
[principles](systemprompts/include/principles.txt) that the prompts ask the model to
follow. Contributions are welcome: open an issue to discuss an idea, or send a pull
request.

You can also change what YACT does without touching the code: every `*.txt` file in
`~/.yact/systemprompts/` becomes a command.

The website in `docs/` is plain HTML, CSS and JavaScript with no build step.
