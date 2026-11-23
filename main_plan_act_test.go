package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yact/commands"
	"yact/config"
)

func readFixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("creating directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// The home directory is replaced so that the system prompts are read from these files
// and never downloaded.
func useTempHome(t *testing.T, plan string, act string, principles string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	promptsDir := filepath.Join(home, ".yact", "systemprompts")
	writeFile(t, filepath.Join(promptsDir, "plan.txt"), plan)
	writeFile(t, filepath.Join(promptsDir, "act.txt"), act)
	writeFile(t, filepath.Join(promptsDir, "include", "principles.txt"), principles)
	return home
}

func enterTempProject(t *testing.T) string {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working directory: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("changing into %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restoring working directory: %v", err)
		}
	})
	return dir
}

// Mirrors runPromptCommand in main.go, without the dispatch in run.
func runPhase(command string, output config.OutputName, client *recordingLLMClient) error {
	promptName, err := config.ParsePromptName(command)
	if err != nil {
		return err
	}

	systemPrompt, err := config.LoadPrompt(promptName)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	userFlags := config.Flaglist{Output: &output}
	resolved := config.ResolveFlags(userFlags, systemPrompt.Defaults)
	return commands.HandleCommand(resolved, cfg, systemPrompt.Text, client.newClientFactory())
}

func TestPlanThenActWritesHelloWorld(t *testing.T) {
	promptBeforePlan := readFixture(t, "test/hello-1/yact-prompt-0.md")
	planReply := readFixture(t, "test/hello-1/yact-prompt-1.md")
	actReply := readFixture(t, "test/hello-1/.yact/temp/buffer.txt")
	planSystemPrompt := readFixture(t, "systemprompts/plan.txt")
	actSystemPrompt := readFixture(t, "systemprompts/act.txt")
	principles := readFixture(t, "systemprompts/include/principles.txt")

	useTempHome(t, planSystemPrompt, actSystemPrompt, principles)
	projectDir := enterTempProject(t)
	promptPath := filepath.Join(projectDir, "yact-prompt.md")
	mainPath := filepath.Join(projectDir, "main.go")
	writeFile(t, promptPath, promptBeforePlan)

	planClient := newRecordingLLMClient(planReply)
	if err := runPhase("plan", config.OutputNamePlan, planClient); err != nil {
		t.Fatalf("plan phase failed: %v", err)
	}
	if planClient.calls != 1 {
		t.Fatalf("plan phase made %d LLM calls, want 1", planClient.calls)
	}
	if got := readFixture(t, promptPath); got != planReply {
		t.Fatalf("yact-prompt.md after the plan phase does not match the planned prompt")
	}

	blockedClient := newRecordingLLMClient(actReply)
	if err := runPhase("act", config.OutputNameApply, blockedClient); err == nil {
		t.Fatalf("act phase succeeded while an assumption was open, want an error")
	}
	if blockedClient.calls != 0 {
		t.Fatalf("blocked act phase made %d LLM calls, want 0", blockedClient.calls)
	}
	if _, err := os.Stat(mainPath); !os.IsNotExist(err) {
		t.Fatalf("main.go exists after the blocked act phase, want it absent")
	}

	planned := readFixture(t, promptPath)
	if !strings.Contains(planned, "- [ ] **A1.") {
		t.Fatalf("planned prompt has no unticked assumption A1 to accept")
	}
	accepted := strings.Replace(planned, "- [ ] **A1.", "- [x] **A1.", 1)
	writeFile(t, promptPath, accepted)

	acceptClient := newRecordingLLMClient(actReply)
	if err := runPhase("act", config.OutputNameApply, acceptClient); err != nil {
		t.Fatalf("act phase failed after the assumption was accepted: %v", err)
	}
	if acceptClient.calls != 1 {
		t.Fatalf("accepted act phase made %d LLM calls, want 1", acceptClient.calls)
	}
	if got := readFixture(t, mainPath); !strings.Contains(got, "func Greeting() string") {
		t.Fatalf("main.go does not contain func Greeting() string")
	}
}
