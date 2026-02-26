package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTSInterfaces_Responses(t *testing.T) {
	t.Parallel()

	f, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "responses.ts"))
	if err != nil {
		t.Fatalf("open responses.ts: %v", err)
	}
	defer f.Close()

	ifaces, err := ParseTSInterfaces(f)
	if err != nil {
		t.Fatalf("parse interfaces: %v", err)
	}
	if got, want := len(ifaces), 45; got != want {
		t.Fatalf("responses interfaces: got %d want %d", got, want)
	}
	if ifaces[0].Name != "AcceptingContactRequest" {
		t.Fatalf("unexpected first interface: %s", ifaces[0].Name)
	}
}

func TestParseTSInterfaces_TypesSnapshot(t *testing.T) {
	t.Parallel()

	f, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "types.ts"))
	if err != nil {
		t.Fatalf("open types.ts: %v", err)
	}
	defer f.Close()

	ifaces, err := ParseTSInterfaces(f)
	if err != nil {
		t.Fatalf("parse types interfaces: %v", err)
	}
	if got := len(ifaces); got < 100 {
		t.Fatalf("types interfaces: got %d want >= 100", got)
	}

	want := map[string]bool{
		"User":      false,
		"Contact":   false,
		"AChatItem": false,
		"Profile":   false,
	}
	for _, iface := range ifaces {
		if _, ok := want[iface.Name]; ok {
			want[iface.Name] = true
		}
	}
	for name, ok := range want {
		if !ok {
			t.Fatalf("types interface %s not parsed", name)
		}
	}
}

func TestRenderTypesRecordsGo(t *testing.T) {
	t.Parallel()

	responses := []TSInterface{
		{
			Name: "ActiveUser",
			Fields: []TSField{
				{Name: "type", TypeExpr: `"activeUser"`},
				{Name: "user", TypeExpr: "T.User"},
			},
		},
	}
	events := []TSInterface{
		{
			Name: "NewChatItems",
			Fields: []TSField{
				{Name: "type", TypeExpr: `"newChatItems"`},
				{Name: "chatItems", TypeExpr: "T.AChatItem[]"},
			},
		},
	}

	src, err := RenderTypesRecordsGo("types", responses, events)
	if err != nil {
		t.Fatalf("render records: %v", err)
	}
	code := string(src)
	if !strings.Contains(code, "type ResponseActiveUser struct") {
		t.Fatalf("missing ResponseActiveUser struct")
	}
	if !strings.Contains(code, "type EventNewChatItems struct") {
		t.Fatalf("missing EventNewChatItems struct")
	}
	if !strings.Contains(code, "func DecodeResponseByType") {
		t.Fatalf("missing DecodeResponseByType")
	}
	if !strings.Contains(code, "func DecodeEventByType") {
		t.Fatalf("missing DecodeEventByType")
	}
}
