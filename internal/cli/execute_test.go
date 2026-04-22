package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"bm/internal/config"
	"bm/internal/testxdg"
)

func TestExecute_successConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"bm", "config", "path"}

	oldOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	oldExit := osExit
	osExit = func(code int) {
		t.Fatalf("unexpected os.Exit(%d)", code)
	}
	defer func() { osExit = oldExit }()

	Execute()

	_ = w.Close()
	os.Stdout = oldOut
	body, _ := io.ReadAll(r)
	_ = r.Close()
	rootCmd.SetArgs(nil)

	if !strings.Contains(string(body), "config.toml") {
		t.Fatalf("stdout %q", body)
	}
}

func TestExecute_errorExits(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"bm", "config", "get", "nope.not.a.key"}

	code := -1
	oldExit := osExit
	osExit = func(c int) { code = c }
	defer func() { osExit = oldExit }()

	Execute()

	rootCmd.SetArgs(nil)

	if code != 1 {
		t.Fatalf("exit code %d", code)
	}
}

func TestRunTUI_stubbedViaRootCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}

	old := tuiRun
	var ran bool
	tuiRun = func() error {
		ran = true
		return nil
	}
	defer func() { tuiRun = old }()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("expected stubbed TUI run")
	}
}
