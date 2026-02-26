package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTSCommands(t *testing.T) {
	t.Parallel()

	commandsMD, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "COMMANDS.md"))
	if err != nil {
		t.Fatalf("open COMMANDS.md: %v", err)
	}
	defer commandsMD.Close()

	catalog, err := ParseCommandsMarkdown(commandsMD)
	if err != nil {
		t.Fatalf("parse commands markdown: %v", err)
	}

	commandsTS, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "commands.ts"))
	if err != nil {
		t.Fatalf("open commands.ts: %v", err)
	}
	defer commandsTS.Close()

	cmds, err := ParseTSCommands(commandsTS, catalog)
	if err != nil {
		t.Fatalf("parse ts commands: %v", err)
	}
	if got, want := len(cmds), 42; got != want {
		t.Fatalf("commands: got %d want %d", got, want)
	}
	if cmds[0].Name != "APICreateMyAddress" {
		t.Fatalf("unexpected first command: %s", cmds[0].Name)
	}
}

func TestRenderCommandRequestsGo(t *testing.T) {
	t.Parallel()

	cmds := []TSCommand{
		{
			Name:   "ShowActiveUser",
			ExprJS: `'/user'`,
		},
		{
			Name: "APICreateMyAddress",
			Fields: []TSField{
				{Name: "userId", TypeExpr: "number", Comment: "int64"},
			},
			ExprJS: `"/_address " + self.userId`,
		},
	}

	src, err := RenderCommandRequestsGo("command", cmds)
	if err != nil {
		t.Fatalf("render command requests: %v", err)
	}
	code := string(src)
	if !strings.Contains(code, "type ShowActiveUser struct") {
		t.Fatalf("missing ShowActiveUser struct")
	}
	if !strings.Contains(code, "type APICreateMyAddress struct") {
		t.Fatalf("missing APICreateMyAddress struct")
	}
	if !strings.Contains(code, "evalCommandExpression") {
		t.Fatalf("missing eval helper")
	}
}
