# Summary

The code is in good shape. Untrusted paths and options are parsed into precise types (`projectpath.Path`, `IgnorePattern`, `ModelName`), pure text logic is tested with table tests, and the Anthropic SDK is confined to one file. The main weaknesses are a few places where logic is mixed with I/O or hides side effects (the LLM response parser, `applyPromptLinks`, `readFile`), loosely typed batch state, and oversized files (`config.go`, `index.go`, `codefile.go`).

## Findings

### F1. Medium: Corrupt batch job file is silently accepted and applied

**Principle:** P7 Parse, don't validate
**Where:** [logic/batch.go](logic/batch.go), `ReadBatchJob`; [commands/batch.go](commands/batch.go), `HandleBatchCommand`
**Problem:** When `json.Unmarshal` fails, the raw file content becomes the batch ID and `Mode` defaults to `OutputApply`. A corrupt or foreign file therefore sends an arbitrary string to the API as an ID. A successful result is then written to disk as code files, because the zero mode is apply. A `Mode` value outside 0 to 2 is also accepted and only fails later with "unknown output mode".
**Fix:** Keep the legacy fallback explicit and narrow. Add `func parseBatchJob(data []byte) (BatchJob, error)`, which accepts either valid JSON with a known mode or a legacy bare ID matching the batch-ID format. In the legacy case it sets the mode explicitly. Return an error for everything else. Validate `Mode` with `func (m OutputMode) Valid() bool` in `config`. Cover this with tests for corrupt, legacy and out-of-range input.

### F2. Medium: Batch results and status use loose strings and exclusive optional fields

**Principle:** P6 Illegal states unrepresentable
**Where:** [api/claude.go](api/claude.go), `BatchResult`, `BatchStatus`, `batchResultFromEntry`; [commands/batch.go](commands/batch.go), `HandleBatchCommand`
**Problem:** `BatchResult` holds `Failure string` next to `Text` and the token fields, so both success and failure data can be set at once, and callers must remember to check `Failure != ""`. `BatchStatus.Ended` duplicates `ProcessingStatus == "ended"`, and the two can disagree. `batchResultFromEntry` switches on the literals "succeeded", "errored", "canceled" and "expired".
**Fix:** Replace `BatchResult` with a sum type: `type BatchOutcome interface{ isBatchOutcome() }`, implemented by `BatchSucceeded{Text string; InputTokens, OutputTokens int64; Cost float64}` and `BatchFailed{Reason string}`. Change `GetBatchResult` to return `(BatchOutcome, error)`. Replace `ProcessingStatus string` and `Ended bool` with `type BatchProcessingState int` (`BatchInProgress`, `BatchEnded`, `BatchCanceling`) set once in `GetBatchStatus`. Use the SDK's result-type constants, or named constants, instead of the literals.

### F3. Medium: Names and signatures hide side effects

**Principle:** P3 No hidden side effects
**Where:** [commands/call_plan.go](commands/call_plan.go), `applyPromptLinks`; [commands/read.go](commands/read.go), `readFile`
**Problem:** `applyPromptLinks` resolves links, scans the project for keyword matches, reads files, prints progress, appends to `requestContent.Files` and rewrites `requestContent.Prompt` (footer and markers stripped). The name says none of this, and a caller cannot tell that `Prompt` changes. `readFile` stats the file, writes `extensions.txt` through `recordReadExtension`, edits `yact-prompt.md` and prints. It also swallows stat errors by printing them and returning nil.
**Fix:** In `call_plan.go`, split into a function that returns data, `func collectPromptFiles(prompt string) (files []logic.CodeFile, unresolved []projectpath.Path, err error)`, and a printing step, `func printReadFiles(files []logic.CodeFile)`. `HandleCommand` then assigns `requestContentForLLM.Files` and `requestContentForLLM.Prompt = sentPromptText(planPrompt)` visibly. In `read.go`, rename `readFile` to `addReadLinkAndExtension(filePath projectpath.Path) error`. Return the stat error, or a typed skip reason, instead of printing it and returning nil.

### F4. Medium: The LLM-response parser reads configuration from disk

**Principle:** P5 Functional core, imperative shell
**Where:** [logic/codefile.go](logic/codefile.go), `ParseCodeFilesDetailed`, `processCodeBlock`, `CodeFile.Apply`
**Problem:** `ParseCodeFilesDetailed` calls `config.LoadExtensions()` inside the parser, so the code that decides which blocks become files written to disk cannot be tested without a project directory. The file also mixes parsing, disk writes (`Apply`, `writeContent`, `removeFile`) and file reading (`ReadAsFile`), and is close to 300 lines.
**Fix:** Change the signature to `func ParseCodeFiles(response string, extensions []string) []ParsedElement`. It cannot fail, so it returns no error. Load the extensions in `commands` (`HandleCommitCommand` and `deliverResponse`) and pass them in. Also compile the five regexes in `extractFilenameFromComment` once, as package-level variables. Then split the file along its responsibilities inside package `logic`: `codefile_parse.go` for the parsing and `codefile_apply.go` for `Apply`, `writeContent`, `removeFile` and `ReadAsFile`.

### F5. Medium: `config.go` has too many jobs and duplicated download logic

**Principle:** P2 Single responsibility (also D4 file size, D1 duplication, P5)
**Where:** [config/config.go](config/config.go), `downloadToFile`, `downloadPromptTree`, `downloadIncludeTree`, `DownloadPrompt`, `LoadExtensions`, `LoadIgnoredFolders`, `LoadPrompt`, `Config`
**Problem:** One file of over 400 lines holds the user config, project paths, prompt loading, include handling, HTTP downloads with console output, and extension and ignore list persistence. `downloadPromptTree` and `downloadIncludeTree` repeat the same "parse include names, validate, recurse" rule. `http.Get` uses the default client with no timeout.
**Fix:** Keep the package public interface unchanged and split into `paths.go` (the `get*` functions), `prompt_download.go`, `project_lists.go` (extensions and ignored folders) and `config.go` (`Config`, `Load`, `Save`, `LoadPrompt`). Merge the two tree functions into `func downloadIncludes(content []byte, visited map[string]bool) error`, used by both. Use `http.Client{Timeout: 30 * time.Second}`. Move the `fmt.Printf` calls from `DownloadPrompt` to `commands.downloadTarget`.

### F6. Medium: `index.go` is oversized and shared helpers live in a language file

**Principle:** P2 Single responsibility (also D3 organization, D4 file size, P16)
**Where:** [index/index.go](index/index.go), [index/indexphp.go](index/indexphp.go) (`matchModifier`, `containsString`, `skipIndentation`, `skipAtAttributes`, `skipModifiers`, `skipOneModifier`, `startsWithKeyword`), [index/bracetree.go](index/bracetree.go)
**Problem:** `index.go` is over 300 lines and mixes language registry, tree building, rendering, signature collection and lexical skipping (`skipLineComment`, `skipBlockComment`, `skipLiteral`, `collapseWhitespace`), which `bracetree.go` uses. Helpers used by every language live in `indexphp.go`, so a reader of the Swift or Kotlin rules must look in the PHP file. `containsString` duplicates `slices.Contains`.
**Fix:** Move the lexical skip functions into `bracetree.go`, or a new `lexer.go`. Move the shared modifier and keyword helpers into `scanhelpers.go`. Move the tree and render functions into `render.go`. Replace `containsString(values, target)` with `slices.Contains(values, target)`. The public interface (`WriteIndex`, `Visibility`) stays unchanged.

### F7. Medium: Cost, request building and LLM-output application are not tested in the input

**Principle:** P15 Tests
**Where:** [api/claude.go](api/claude.go), `calculateCost`, `selectTier`; [api/client.go](api/client.go), `Serialize`, `BuildUserBlocks`, `FormatCost`; [commands/commit_test.go](commands/commit_test.go), `codeFileBlock` and `fence`; [commands/call_deliver.go](commands/call_deliver.go), `processResponse`
**Problem:** The input has no test files for the `api` package. The helpers `codeFileBlock` and `fence` in `commit_test.go` are defined but never used, and the one test there covers only a text element. Cost calculation (the Haiku long-prompt threshold, the batch multiplier), the block order sent to the model (`domain`, `index`, `files`, `prompt`) and parsing LLM output that is written to disk are the most consequential pure logic. If tests exist elsewhere, ignore this finding.
**Fix:** Add `api/client_test.go` with tests for `Serialize` (the `.sh` delimiter, the header line) and `BuildUserBlocks` (order, skipping empty blocks). Add `api/cost_test.go` for `ClaudeClient.calculateCost` and `calculateBatchCost` at, below and above `haikuLongPromptThreshold`. Add `logic/codefile_test.go` for `ParseCodeFiles`: truncated final block, unlisted extension, `..` path, empty content meaning delete. Use `codeFileBlock` in the `processResponse` tests, or delete it.

### F8. Low: `ClaudeClient` needs a second step to become usable

**Principle:** P8 Explicit over implicit
**Where:** [api/claude.go](api/claude.go), `ClaudeClient.Init`; [api/client.go](api/client.go), `Client` interface; [commands/call_request.go](commands/call_request.go), `newClient`
**Problem:** `&api.ClaudeClient{}` is valid but unusable until `Init` is called, and `Init` is part of the interface. Forgetting it gives an empty API key and a late error. `selectTier` also uses `longInputPrice: 0` to mean "no long-prompt pricing", which is implicit.
**Fix:** Add `func NewClaudeClient(cfg *config.Config) *ClaudeClient` and delete `Init` from both the type and the `Client` interface. `newClient` becomes `return api.NewClaudeClient(cfg)`. Replace the zero-price convention with `longPrompt *promptPricing` (nil means none), or add `hasLongPricing bool`.

### F9. Low: Misleading names and an error that is never returned

**Principle:** P1 Searchable names (also P13)
**Where:** [api/client.go](api/client.go), `FormatCost`, `BuildUserBlocks`; [api/claude.go](api/claude.go), `buildParams`
**Problem:** `FormatCost` computes a cost and formats nothing. `BuildUserBlocks` always returns a nil error, which forces dead error handling in `buildParams`.
**Fix:** Rename `FormatCost` to `CalculateTokenCost`. Change `BuildUserBlocks` to `func BuildUserBlocks(requestContent *logic.RequestContent) []string`. Change `buildParams` to return `anthropic.MessageNewParams` without an error, and update `Call` and `SubmitBatch` accordingly.

### F10. Low: Two types for one concept, output mode

**Principle:** P1 Searchable names (one name per concept; also P13)
**Where:** [config/flaglist.go](config/flaglist.go), `OutputMode`, `ParseOutputMode`, `Flaglist.EffectiveOutput`; [config/option_values.go](config/option_values.go), `OutputName`; [logic/batch.go](logic/batch.go), `BatchJob.Mode`
**Problem:** The output option exists as the string type `OutputName` (parsed from the CLI and prompt files) and as the integer type `OutputMode` (used in commands and persisted in the batch job). `ParseOutputMode(flags Flaglist)` does not parse anything, it converts. `Flaglist.Output` is a pointer that `ResolveFlags` always fills, yet `EffectiveOutput` still handles nil.
**Fix:** Use `OutputName` as the single type, with `switch` on its constants in `deliverResponse`. Store it in `BatchJob.Mode` as a string. Keep reading legacy integers in `parseBatchJob` from F1 and map 0, 1 and 2 to the matching names. Alternatively, keep `OutputMode` and rename `ParseOutputMode` to `OutputModeOf(flags Flaglist) OutputMode`. Change `ResolveFlags` to return a type with a non-pointer `Output` so the nil case disappears.

## Strengths

- P7 and P6: `projectpath.Parse` rejects absolute and escaping paths once at the boundary, `ParseIgnorePattern` keeps its fields unexported so every value is valid, and `PromptGate` and `ApplyOutcome` model states as enums instead of booleans.
- P5 and P15: the prompt text logic (`SplitPromptFooter`, `CarryAssumptionTicks`, `DecidePromptGate`, `CompactPromptLinks`, `ContainsKeyword`) is pure and covered by table-driven tests without mocks.
- P10 and P14: the Anthropic SDK is used only in [api/claude.go](api/claude.go) behind the `Client` interface, and comments such as the one on `batchResult` storage order in `HandleBatchCommand` explain why.

## Not assessed

- The `HandleIndexCommand` and `HandleKeywordCommand` functions index `args[0]` without a length check. Whether that is safe depends on the dispatcher, which is not in the input (for example `main.go`).
- `printWarnings`, `printSentPrompt`, `GlobPattern` and `ParseGlobPattern` appear only in tests in the input. Whether they are dead code depends on callers outside the input.