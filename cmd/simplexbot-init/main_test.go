package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultNameFromModule(t *testing.T) {
	tests := []struct {
		module string
		want   string
	}{
		{module: "github.com/acme/my-bot", want: "My Bot"},
		{module: "github.com/acme/my_bot", want: "My Bot"},
		{module: "mybot", want: "Mybot"},
	}

	for _, tc := range tests {
		if got := defaultNameFromModule(tc.module); got != tc.want {
			t.Fatalf("defaultNameFromModule(%q) = %q, want %q", tc.module, got, tc.want)
		}
	}
}

func TestBuildProjectFiles_Basic(t *testing.T) {
	files, err := buildProjectFiles(projectConfig{
		Module:    "github.com/acme/my-bot",
		Name:      "My Bot",
		WSURL:     "wss://bot.example/ws",
		SDKModule: "github.com/Malomalsky/go-simplex",
		Template:  projectTemplateBasic,
	})
	if err != nil {
		t.Fatalf("buildProjectFiles returned error: %v", err)
	}

	mustContain(t, files["go.mod"], "module github.com/acme/my-bot")
	mustContain(t, files["main.go"], "github.com/Malomalsky/go-simplex/sdk/bot")
	mustContain(t, files["main.go"], "wss://bot.example/ws")
	mustContain(t, files["main.go"], "/echo <text>")
	mustContain(t, files["README.md"], "Bot API commands")
	mustContain(t, files[".gitignore"], ".codex/")
}

func TestBuildProjectFiles_Moderation(t *testing.T) {
	files, err := buildProjectFiles(projectConfig{
		Module:    "github.com/acme/mod-bot",
		Name:      "Mod Bot",
		WSURL:     "ws://localhost:5225",
		SDKModule: "github.com/Malomalsky/go-simplex",
		Template:  projectTemplateModeration,
	})
	if err != nil {
		t.Fatalf("buildProjectFiles returned error: %v", err)
	}

	mustContain(t, files["main.go"], "type denyList struct")
	mustContain(t, files["main.go"], "/addword <word>")
	mustContain(t, files["README.md"], "moderation template")
	mustContain(t, files["README.md"], "`/words`")
}

func TestBuildProjectFiles_InvalidTemplate(t *testing.T) {
	_, err := buildProjectFiles(projectConfig{Template: "unknown"})
	if err == nil {
		t.Fatalf("expected error for invalid template")
	}
}

func TestWriteProject_NoForceDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"go.mod": "module test\n",
	}

	if err := writeProject(dir, files, false); err != nil {
		t.Fatalf("first writeProject returned error: %v", err)
	}
	if err := writeProject(dir, files, false); err == nil {
		t.Fatalf("second writeProject expected overwrite error")
	}
}

func TestWriteProject_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte("module old\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	files := map[string]string{
		"go.mod": "module new\n",
	}
	if err := writeProject(dir, files, true); err != nil {
		t.Fatalf("writeProject(force) returned error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if string(got) != "module new\n" {
		t.Fatalf("go.mod = %q, want %q", string(got), "module new\\n")
	}
}

func mustContain(t *testing.T, text, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("text does not contain %q", want)
	}
}
