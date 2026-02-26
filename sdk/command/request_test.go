package command

import "testing"

func TestCommandStrings(t *testing.T) {
	t.Parallel()

	if got, want := (ShowActiveUser{}).CommandString(), "/user"; got != want {
		t.Fatalf("ShowActiveUser string: got %q want %q", got, want)
	}
	if got, want := (APIShowMyAddress{UserID: 5}).CommandString(), "/_show_address 5"; got != want {
		t.Fatalf("APIShowMyAddress string: got %q want %q", got, want)
	}
	if got, want := (APICreateMyAddress{UserID: 9}).CommandString(), "/_address 9"; got != want {
		t.Fatalf("APICreateMyAddress string: got %q want %q", got, want)
	}
	if got, want := (APIDeleteMyAddress{UserID: 7}).CommandString(), "/_delete_address 7"; got != want {
		t.Fatalf("APIDeleteMyAddress string: got %q want %q", got, want)
	}
}

func TestSendMessagesString(t *testing.T) {
	t.Parallel()

	ttl := 30
	cmd := APISendMessages{
		SendRef:          "@42",
		LiveMessage:      true,
		TTL:              &ttl,
		ComposedMessages: []any{map[string]any{"msgContent": map[string]any{"type": "text", "text": "hi"}, "mentions": map[string]any{}}},
	}
	got := cmd.CommandString()
	want := "/_send @42 live=on ttl=30 json [{\"mentions\":{},\"msgContent\":{\"text\":\"hi\",\"type\":\"text\"}}]"
	if got != want {
		t.Fatalf("APISendMessages string mismatch:\n got: %s\nwant: %s", got, want)
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()

	def, ok := Lookup("APISendMessages")
	if !ok {
		t.Fatalf("expected APISendMessages definition")
	}
	if def.Name != "APISendMessages" {
		t.Fatalf("unexpected definition: %+v", def)
	}
	if _, ok := Lookup("DoesNotExist"); ok {
		t.Fatalf("unexpected definition for unknown command")
	}
}
