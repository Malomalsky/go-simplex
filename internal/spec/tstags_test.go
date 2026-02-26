package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTaggedInterfaces_ResponsesSnapshot(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "spec", "upstream", "responses.ts")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open responses snapshot: %v", err)
	}
	defer f.Close()

	tags, err := ParseTaggedInterfaces(f)
	if err != nil {
		t.Fatalf("parse response interfaces: %v", err)
	}
	if got, want := len(tags), 45; got != want {
		t.Fatalf("response tags: got %d want %d", got, want)
	}

	first := tags[0]
	if first.Name != "AcceptingContactRequest" || first.Tag != "acceptingContactRequest" {
		t.Fatalf("unexpected first response: %+v", first)
	}
}

func TestParseTaggedInterfaces_EventsSnapshot(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "spec", "upstream", "events.ts")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open events snapshot: %v", err)
	}
	defer f.Close()

	tags, err := ParseTaggedInterfaces(f)
	if err != nil {
		t.Fatalf("parse event interfaces: %v", err)
	}
	if got, want := len(tags), 45; got != want {
		t.Fatalf("event tags: got %d want %d", got, want)
	}

	first := tags[0]
	if first.Name != "ContactConnected" || first.Tag != "contactConnected" {
		t.Fatalf("unexpected first event: %+v", first)
	}
}

func TestRenderTypesTagsGo(t *testing.T) {
	t.Parallel()

	responses := []TaggedType{
		{Name: "ActiveUser", Tag: "activeUser"},
	}
	events := []TaggedType{
		{Name: "NewChatItems", Tag: "newChatItems"},
	}

	src, err := RenderTypesTagsGo("types", responses, events)
	if err != nil {
		t.Fatalf("render tags: %v", err)
	}

	code := string(src)
	if !strings.Contains(code, `ResponseTypeActiveUser ResponseType = "activeUser"`) {
		t.Fatalf("missing response constant:\n%s", code)
	}
	if !strings.Contains(code, `EventTypeNewChatItems EventType = "newChatItems"`) {
		t.Fatalf("missing event constant:\n%s", code)
	}
}
