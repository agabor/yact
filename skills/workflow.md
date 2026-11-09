---
name: yact-workflows
description: Write or edit GitHub Actions workflows that run the YACT CLI (y) to plan, implement or review issues and pull requests. Use when adding or changing a yact-*.yml workflow.
---

# Writing YACT GitHub workflows

YACT (`y`) makes one LLM call per prompt command (`plan`, `act`, `review`, `analyze`, `ask`) and keeps all state in files, so a workflow is a short, linear script. `.github/workflows/` holds three references. Read the closest one before writing, and keep its input names (`issue` or `pr`, and `model`) and its step order unless asked otherwise:

- `yact-issue.yml`: issue, then plan, code, tests and a pull request
- `yact-pr.yml`: PR plan plus the latest comment, then an updated plan and code pushed to the PR branch
- `yact-review.yml`: PR diff, plan and original request, then a review comment

Those files can lag behind the rules below. Where they differ, follow this skill.

## How `y` behaves in CI

- **State is files.** `yact-prompt.md` holds the task or plan. `.yact/temp/` holds `buffer.txt` (the last response, written after every call), `index.txt` and `output.txt`. Both are gitignored, so `git add -A` does not commit them. `y new` clears the task.
- **Prompt.** With no argument, `yact-prompt.md` is the prompt and the files linked in it are sent. An argument is appended after its text. `y read <file>` adds a link. `y output < file` sends stdin, such as a diff, before the files.
- **`y plan`** writes the response to `yact-prompt.md` and changes no code. On an existing plan it updates it. If the plan lists files under `## Missing context`, run `y plan` once more: the linked files are sent automatically. When nothing is left to do the plan reads `No steps remain.` and has no steps.
- **`y act`** implements the unticked steps, writes the files and ticks the steps in `yact-prompt.md`. It exits 1 while the plan has `## Missing context`, and when every step is ticked.
- **Print-only prompts** (`review`, `analyze`, `ask`, or any call with `-n`) still save the response to `.yact/temp/buffer.txt`. Post that file; do not scrape stdout.
- **Always pass** `--no-progress` and `-m "$MODEL"` (`haiku`, `sonnet`, `opus` or `fable`). Default to `haiku` for implementing and `sonnet` for reviewing.
- **API key.** `y config anthropic_api_key <key>` is required. `y` does not read the `ANTHROPIC_API_KEY` variable. The key is stored in `~/.yact/config`.
- **System prompts** are downloaded on first use from the `main` branch of the YACT repository. The runner needs access to raw.githubusercontent.com, and the prompts can be newer than the installed binary.
- **Debug** without spending tokens by adding `--dry-run --verbose`: it prints the files and prompt that would be sent and skips the call.

## Template: issue to pull request

```yaml
name: yact issue

on:
  workflow_dispatch:
    inputs:
      issue:
        description: Issue number to implement
        required: true
        type: string
      model:
        description: Claude model
        type: choice
        options: [haiku, sonnet, opus, fable]
        default: haiku

permissions:
  contents: write
  issues: read
  pull-requests: write

jobs:
  implement:
    runs-on: ubuntu-latest
    env:
      ISSUE_NUMBER: ${{ inputs.issue }}
      MODEL: ${{ inputs.model }}
      TRIGGERING_ACTOR: ${{ github.triggering_actor }}
      GH_TOKEN: ${{ github.token }}
      GH_REPO: ${{ github.repository }}
    steps:
      - name: Validate inputs
        run: |
          case "$ISSUE_NUMBER" in ''|*[!0-9]*) echo "issue must be a number" >&2; exit 1 ;; esac

      - uses: actions/checkout@v7

      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod

      - name: Install yact
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          mkdir -p "$HOME/bin"
          curl -fsSL -o "$HOME/bin/y" https://github.com/agabor/yact/releases/latest/download/yact-linux-amd64
          chmod +x "$HOME/bin/y"
          echo "$HOME/bin" >> "$GITHUB_PATH"
          "$HOME/bin/y" config anthropic_api_key "$ANTHROPIC_API_KEY"

      - name: Configure git identity
        run: |
          user_id=$(gh api "users/$TRIGGERING_ACTOR" --jq '.id')
          git config user.name "$TRIGGERING_ACTOR"
          git config user.email "$user_id+$TRIGGERING_ACTOR@users.noreply.github.com"

      - name: Fetch issue
        run: |
          y new
          # Issue text is untrusted: it only travels through this file.
          gh issue view "$ISSUE_NUMBER" --json title,body --jq '.title + "\n\n" + (.body // "")' > yact-prompt.md

      - name: Create plan
        run: y -m "$MODEL" --no-progress plan

      - name: Refine plan with missing context
        run: |
          if grep -qxF '## Missing context' yact-prompt.md; then
            y -m "$MODEL" --no-progress plan
          fi

      - name: Check plan has no missing context
        run: |
          if grep -qxF '## Missing context' yact-prompt.md; then
            echo "Plan still has missing context" >&2
            cat yact-prompt.md >&2
            exit 1
          fi

      - name: Create branch
        run: git checkout -b "yact/issue-$ISSUE_NUMBER"

      - name: Implement plan
        id: implement
        run: y -m "$MODEL" --no-progress act

      - name: Upload buffer
        if: always() && steps.implement.outcome == 'failure'
        uses: actions/upload-artifact@v7
        with:
          name: yact-buffer-issue-${{ inputs.issue }}
          path: .yact/temp/buffer.txt
          if-no-files-found: ignore

      - name: Remove API key before running generated code
        run: rm -f "$HOME/.yact/config"

      - name: Commit and push
        run: |
          git add -A
          if git diff --cached --quiet; then
            echo "No changes to commit"
          else
            git commit -m "Implement yact plan for issue #$ISSUE_NUMBER"
          fi
          git push -u origin "yact/issue-$ISSUE_NUMBER"

      - name: Run tests
        run: go test ./...

      - name: Open pull request
        run: |
          {
            cat yact-prompt.md
            echo
            echo "Closes #$ISSUE_NUMBER"
          } > "$RUNNER_TEMP/pr-body.md"
          gh pr create \
            --head "yact/issue-$ISSUE_NUMBER" \
            --title "Implement issue #$ISSUE_NUMBER" \
            --body-file "$RUNNER_TEMP/pr-body.md"
```

Why the order is what it is: the plan is refined once and then checked, so a plan that still needs files fails before any code is written. The plan is posted as the PR body after `act`, so ticked boxes show what was done, and later workflows read that body as the plan. The branch is pushed before the tests run, so a test failure leaves the work reviewable.

## Variations

**PR follow-up** (input `pr`, no new branch, no PR created)
- After `actions/checkout`, run `gh pr checkout "$PR_NUMBER"`.
- Build `yact-prompt.md` from the PR body (the plan, with ticked steps) plus one trusted comment, then run `y plan`. It drops finished steps and adds the ones the comment asks for. Then refine and check as in the template.
- Stop successfully, skipping `act`, commit and tests, when `grep -qxF 'No steps remain.' yact-prompt.md`. Do not run `act` on that plan.
- Select the comment by author, and skip the step when none qualifies:
  ```bash
  gh pr view "$PR_NUMBER" --json comments \
    --jq '[.comments[] | select(.authorAssociation | IN("OWNER","MEMBER","COLLABORATOR"))] | .[-1].body // empty'
  ```
  Check that the comments your own review workflow posts still pass the filter, or add that bot's login to it.
- Finish with `git push`.

**Review** (input `pr`, read-only)
- `permissions:` are `contents: read`, `issues: read`, `pull-requests: write`. No toolchain, no git identity.
- Plan: `gh pr view "$PR_NUMBER" --json body --jq '.body // ""' > yact-prompt.md`.
- Original request: for each of `closingIssuesReferences`, the issue title and body; if there are none, the PR title. Write it to `$RUNNER_TEMP/request.txt`.
- Context: `gh pr diff "$PR_NUMBER" > "$RUNNER_TEMP/pr.diff"` then `y output < "$RUNNER_TEMP/pr.diff"`. Then `y read "$file"` for each changed file that still exists and is text (`[ -f "$file" ] && grep -Iq . "$file"`).
- Run `y -m "$MODEL" --no-progress review "$(cat "$RUNNER_TEMP/request.txt")"`, then `gh pr comment "$PR_NUMBER" --body-file .yact/temp/buffer.txt`.
- Never build or run the PR's code in this workflow. That is what makes it safe to review pull requests from forks.

**Other prompts** (`analyze`, `ask`): the same install step, context from `y read` or `y keyword`, a print-only call, and the buffer posted as a comment or uploaded as an artifact.

## Rules

1. **Secrets and inputs go through `env:`, and are quoted in the script** (`"$VAR"`). Never put `${{ ... }}` inside `run:` text. `secrets.*` appears only in `env:` or `with:`. actionlint reports violations as SC2086.
2. **Validate number inputs** before first use, as the template does.
3. **Issue text, PR bodies, comments, diffs and file names are untrusted.** They move only through files and quoted variables. Never put them into `run:` text, `eval`, `$GITHUB_ENV` or `$GITHUB_OUTPUT`.
4. **Only trusted people may steer a job that can write.** Anyone can comment on a public PR, so filter comments by author (see the PR follow-up) and never take "the last comment" from an arbitrary author.
5. **Give each workflow the least `permissions:` it needs.** Review: `contents: read`. Implement: `contents: write` and `pull-requests: write`. Add `issues: read` only when issues are read.
6. **Generated code runs without the API key.** Delete `~/.yact/config` right after `act`, before any test or build, as the template does. For stronger isolation run the tests in a second job (`needs:`) that has `contents: read` and no secrets.
7. **Keep `workflow_dispatch` as the trigger.** For `issues`, `issue_comment` or `pull_request_target`, check the author's association first, and never run fork code in a job that holds the key.
8. **Pin YACT when runs must be reproducible:** replace `releases/latest/download/` with `releases/download/<tag>/`.
9. **Fail loudly.** Upload `.yact/temp/buffer.txt` when `act` fails, and print the plan to stderr when it is rejected.
10. **Finish with `actionlint`.** Name the file `yact-<purpose>.yml`, match `name:` to it, and check that every `y` call has `--no-progress` and `-m "$MODEL"`.