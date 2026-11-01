# Goal

When a command is called with a prompt argument, write the combined text of `yact-prompt.md` and that argument to `yact-prompt.md` before the LLM call, so the file matches the prompt that was sent. Later calls with a prompt read this file again, so the combined text grows with each call.

## Steps

### [x] 1. Add the function that writes the project prompt file

**File:** [logic/project_prompt.go](logic/project_prompt.go) (new)

- Add `func WriteProjectPrompt(text string) error`, which writes `text` to the path returned by `config.GetProjectPromptPath()`, replacing the previous content
- Keep the file write in this function only, so the caller passes the text and the function does no string handling (P5: I/O in the outer layer)

**Context:** none

### [x] 2. Write the combined prompt before the LLM call

**File:** [commands/call.go](commands/call.go)

- Declare `combinedPrompt := ""` next to `planPrompt := ""`
- In the branch without a plan step selection, assign `combinedPrompt = combinePromptText(promptFileText, prompt)` and set `requestContentForLLM.Prompt` to the same value
- Keep `combinedPrompt` as the raw text, before `applyPromptLinks` expands links, so the file holds the prompt and not the expanded file contents
- After the dry-run check and before the batch and LLM calls, when `planUsed` is false, call `func WriteProjectPrompt(text string) error` with `combinedPrompt` and return its error (P3: the write is an explicit, visible effect)
- Do not write the file when a plan is used, because the plan content is not the prompt file
- Leave the dry run without a write, so a dry run does not change the prompt file

**Context:** [logic/project_prompt.go](logic/project_prompt.go)