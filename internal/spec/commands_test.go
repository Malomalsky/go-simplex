package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCommandsMarkdown_UpstreamSnapshot(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "spec", "upstream", "COMMANDS.md")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	doc, err := ParseCommandsMarkdown(f)
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	if got, want := len(doc.Categories), 8; got != want {
		t.Fatalf("categories: got %d want %d", got, want)
	}

	all := doc.AllCommands()
	if got, want := len(all), 42; got != want {
		t.Fatalf("commands: got %d want %d", got, want)
	}

	if all[0].Name != "APICreateMyAddress" {
		t.Fatalf("unexpected first command: %q", all[0].Name)
	}
	if all[0].NetworkUsage != NetworkUsageInteractive {
		t.Fatalf("unexpected network usage for %s: %q", all[0].Name, all[0].NetworkUsage)
	}

	send := findByName(all, "APISendMessages")
	if send == nil {
		t.Fatalf("APISendMessages not found")
	}
	if send.Syntax == "" {
		t.Fatalf("APISendMessages syntax is empty")
	}
	if send.NetworkUsage != NetworkUsageBackground {
		t.Fatalf("APISendMessages network usage: got %q want %q", send.NetworkUsage, NetworkUsageBackground)
	}
	if len(send.Params) == 0 {
		t.Fatalf("APISendMessages should have parameters")
	}

	for _, c := range all {
		if c.Syntax == "" {
			t.Fatalf("command %s has empty syntax", c.Name)
		}
	}
}

func TestRenderCatalogGo(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "spec", "upstream", "COMMANDS.md")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	doc, err := ParseCommandsMarkdown(f)
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	rendered, err := RenderCatalogGo(doc, "command")
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	src := string(rendered)
	if !strings.Contains(src, "var GeneratedCatalog = []Definition{") {
		t.Fatalf("generated source missing catalog declaration")
	}
	if !strings.Contains(src, `"APICreateMyAddress"`) {
		t.Fatalf("generated source missing APICreateMyAddress")
	}
	if !strings.Contains(src, `"APISetContactPrefs"`) {
		t.Fatalf("generated source missing APISetContactPrefs")
	}
}

func findByName(commands []CommandSpec, name string) *CommandSpec {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}
