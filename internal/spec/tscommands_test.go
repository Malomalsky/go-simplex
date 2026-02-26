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
			ExprJS: `'/_address ' + self.userId`,
		},
		{
			Name: "ReceiveFile",
			Fields: []TSField{
				{Name: "fileId", TypeExpr: "number", Comment: "int64"},
				{Name: "storeEncrypted", Optional: true, TypeExpr: "boolean"},
			},
			ExprJS: `'/freceive ' + self.fileId + (typeof self.storeEncrypted == 'boolean' ? ' encrypt=' + (self.storeEncrypted ? 'on' : 'off') : '')`,
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
	if !strings.Contains(code, `return "/user"`) {
		t.Fatalf("missing direct string rendering")
	}
	if !strings.Contains(code, `jsToString(c.UserId)`) {
		t.Fatalf("missing typed field rendering")
	}
	if !strings.Contains(code, `jsLooseEqual(jsTypeOf(c.StoreEncrypted), "boolean")`) {
		t.Fatalf("missing typeof/equality rendering")
	}
}
