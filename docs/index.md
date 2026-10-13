# YACT: plan-driven, scriptable LLM edits from the command line

**Source code:** [github.com/agabor/yact](https://github.com/agabor/yact)

**Author:** Gábor Angyal

**License:** GNU GPL v3, see [LICENSE](https://github.com/agabor/yact/blob/main/LICENSE) in the repository

YACT (Yet Another Coding Tool) is a small, open-source CLI that makes one LLM call per command and writes complete source files back to your project. It does not explore your codebase, run commands, or loop on its own. You decide what the model sees and what it is asked to do.

It is not trying to be a full coding agent, and it is not a replacement for tools like aider, Claude Code or Cursor. It is a sharp, narrow tool for developers who want AI-assisted edits they can predict, script and review.

## Who it is for

YACT is a good fit if you:

- know your codebase well enough to pick the relevant files yourself
- want to see exactly what is sent to the model, and what it costs, on every call
- prefer composing Unix commands to chatting with an agent
- keep source files small and well factored
- want AI edits to fit into scripts, Makefiles or batch jobs

It is probably not for you if you want an assistant that finds the right files on its own, runs your tests and fixes what broke, or works across a large, unfamiliar codebase.

## What makes it different

### Plans are files you can edit, review and resume

`y -p plan "add authentication"` asks the model for a short Markdown plan and saves it to `.yact/plan.md`. Each step changes exactly one file and ends with links to the files that step needs:

```markdown
### [ ] 2. Add token check to the router
...
**Context:** [src/router/router.go](src/router/router.go), [src/auth/token.go](src/auth/token.go)
```

You can read and edit the plan before anything is written. Then run it one step at a time:

```
y -s next act
```

Each step sends only its own linked files, so calls stay small and focused. Finished steps are ticked off in the file, so you can stop, review the diff, fix the plan, and pick up where you left off. The plan is plain Markdown, so you can commit it, put it in a pull request, or hand it to a colleague.

### Every command is a single, stateless call

There is no session to manage. `y act "add a --version flag"` reads the context, makes one API call, writes the result and exits. That makes YACT easy to drive from a shell loop, a Makefile target or a CI job, and easy to reason about: one command, one request, one response.

### Prompts are commands

Every text file in `~/.yact/systemprompts/` becomes a command. `act`, `ask`, `plan` and `bash` ship as official prompts, but they are just text files you can read and change. Drop in `review.txt` and you have `y review`. A `flags:` line at the top of a prompt sets its defaults, such as the model or whether it may write files.

### Transparent by design

- The task context is a plain text list of file paths in `.yact/context.txt`.
- Every raw response is logged to `.yact/buffer.txt`.
- Every call prints the model, duration, input and output tokens, and an estimated cost.
- A configurable input-line limit stops you from accidentally sending far more than you meant to.
- There is no telemetry and no background process. The only network calls are the ones you trigger.

### Small enough to audit

YACT is a single Go binary built from roughly 2,600 lines of code, licensed under GPL v3. You can read the whole thing in an afternoon, and building from source is one command.

## A typical session

```
y new                          # fresh task context
y read "internal/billing/*.go" # pick the files
y ask "why does invoice rounding drift?"
y -p plan "fix rounding to use integer cents"
$EDITOR .yact/plan.md          # adjust the plan if needed
y -s next act                  # implement step 1
git diff                       # review
y -s next act                  # step 2, and so on
```

## Honest trade-offs

YACT makes deliberate choices, and each has a cost:

- **You pick the context.** YACT never searches for the files a task needs. If you leave out something relevant, the model will not know about it.
- **Whole files, not diffs.** The model rewrites complete files. This keeps the interface simple and avoids broken patches, but every edited file must fit within the output token limit, and rewriting a long file to change one line is slower and costs more. Files of a few hundred lines work best.
- **No feedback loop.** YACT does not run your tests or linter and retry. You run them and decide what to do next.
- **Limited providers.** Today it supports Anthropic's API and Qwen models on AWS Bedrock. There is no support yet for local models or other providers.
- **It writes to disk.** Run it in a version-controlled working tree and review changes with `git diff`, or use `y -n` to print a response and `y commit` to apply it after you've read it.
- **It is young.** YACT is a small project with one maintainer, no published benchmarks yet, and no automated test suite. Expect rough edges.

## How it compares

| | YACT | Full coding agents (aider, Claude Code, Cursor, ...) |
| --- | --- | --- |
| Interaction | One-shot CLI commands | Interactive session or chat |
| Context | Chosen by you | Partly or fully chosen by the tool |
| API calls per command | Exactly one | Often many |
| Edit format | Complete files | Diffs, patches or tool calls |
| Planning | Editable Markdown plan with per-step context | Varies; usually inside the session |
| Runs tests, retries | No | Often yes |
| Model choice | Anthropic, Qwen on Bedrock | Usually much broader |
| Size | ~2,600 lines of Go | Much larger |

If you want a tool that does more on its own, use one of those. If you want one that does exactly what you ask, one call at a time, and leaves a plan you can read, try YACT.

## Get started

```
git clone https://github.com/agabor/yact.git
cd yact
go build -o y .
mkdir -p ~/.local/bin && mv y ~/.local/bin/

y config anthropic_api_key <your-own-anthropic-key>
y -d act
y -d plan
y -d ask
```

Prebuilt binaries for Linux, macOS and Windows are on the [releases page](https://github.com/agabor/yact/releases).

## Contribute

YACT is GPL v3 and welcomes issues and pull requests at [github.com/agabor/yact](https://github.com/agabor/yact). Useful places to help right now: tests for the response parser and plan logic, support for OpenAI-compatible endpoints, and new system prompts.