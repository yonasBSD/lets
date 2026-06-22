package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestZshActiveCommandStopsAtDoubleDash(t *testing.T) {
	output := runZshCompletionHelper(t, `
words=(-- foo)
_lets_active_command
`)

	if output != "" {
		t.Fatalf("expected no active command after --, got %q", output)
	}
}

func TestZshCheckLetsConfigRejectsCommandTokens(t *testing.T) {
	output := runZshCompletionHelper(t, `
fake_lets() { return 0 }
LETS_EXECUTABLE=fake_lets
_check_lets_config foo
`)

	if output != "1" {
		t.Fatalf("expected command token to be rejected, got %q", output)
	}
}

func runZshCompletionHelper(t *testing.T, body string) string {
	t.Helper()

	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh is not available")
	}

	var completion bytes.Buffer
	if err := genZshCompletion(&completion); err != nil {
		t.Fatalf("generate zsh completion: %v", err)
	}

	script := `
compdef() { : }
compinit() { : }
` + completion.String() + body

	cmd := exec.Command("zsh", "-fc", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zsh completion helper failed: %v\n%s", err, out)
	}

	return strings.TrimSpace(string(out))
}
