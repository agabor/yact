# Goal

Inline the `WriteProjectPrompt` helper from `logic/project_prompt.go` into the `HandleCommand` function in `commands/call.go`, then delete the now-unused file. The project prompt is still written to `.yact/prompt.md` at the same point in the flow, with the same content and error handling.

## Steps

### [ ] 1. Write the project prompt directly in the command handler

**File:** [commands/call.go](commands/call.go)

- In `func HandleCommand(flags config.Flaglist, cfg *config.Config, systemPrompt string, prompt string) error`, replace the call to `logic.WriteProjectPrompt(combinedPrompt)` with `os.WriteFile(config.GetProjectPromptPath(), []byte(combinedPrompt), 0644)`, keeping the same `if err != nil { return err }` handling.
- Add the `"os"` import. The `logic` import stays, because other calls still use it.
- Keep the existing comment above the write, since it explains why the file must hold the sent prompt (P14).

**Context:** [logic/project_prompt.go](logic/project_prompt.go)

### [ ] 2. Delete the project prompt helper file

**File:** [logic/project_prompt.go](logic/project_prompt.go)

- Delete the file. Its only function, `WriteProjectPrompt`, is now written inline in `HandleCommand`, so no code references it (P13: no unused code).

**Context:** none