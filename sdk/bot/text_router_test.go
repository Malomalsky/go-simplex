package bot

import (
	"context"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
)

func TestTextRouterHandleKnownCommand(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	got := make(chan TextCommand, 1)
	if err := router.On("start", func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		got <- cmd
		return nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}

	err := router.Handle(context.Background(), nil, DirectTextMessage{
		ContactID: 42,
		Text:      "/start hello world",
	})
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}

	select {
	case cmd := <-got:
		if cmd.Name != "start" {
			t.Fatalf("unexpected command name: %s", cmd.Name)
		}
		if cmd.Args != "hello world" {
			t.Fatalf("unexpected command args: %q", cmd.Args)
		}
		if cmd.Message.ContactID != 42 {
			t.Fatalf("unexpected contact id: %d", cmd.Message.ContactID)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting command handler")
	}
}

func TestTextRouterUnknownCommand(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	called := make(chan string, 1)
	router.OnUnknown(func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		called <- cmd.Name
		return nil
	})

	err := router.Handle(context.Background(), nil, DirectTextMessage{Text: "/missing arg"})
	if err != nil {
		t.Fatalf("handle unknown command: %v", err)
	}

	select {
	case name := <-called:
		if name != "missing" {
			t.Fatalf("unexpected unknown command name: %s", name)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting unknown command handler")
	}
}

func TestTextRouterCaseInsensitive(t *testing.T) {
	t.Parallel()

	router := NewTextRouter(WithCommandCaseInsensitive(true))
	got := make(chan TextCommand, 1)
	if err := router.On("HeLp", func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		got <- cmd
		return nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}

	if err := router.Handle(context.Background(), nil, DirectTextMessage{Text: "/HELP now"}); err != nil {
		t.Fatalf("handle command: %v", err)
	}

	select {
	case cmd := <-got:
		if cmd.Name != "help" {
			t.Fatalf("expected normalized lower command, got: %s", cmd.Name)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting command handler")
	}
}

func TestTextRouterRequirePrefix(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	called := false
	if err := router.On("start", func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}

	if err := router.Handle(context.Background(), nil, DirectTextMessage{Text: "start"}); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if called {
		t.Fatalf("expected message without prefix to be ignored")
	}
}

func TestTextRouterOnValidation(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	if err := router.On("bad command", func(context.Context, *client.Client, TextCommand) error { return nil }); err == nil {
		t.Fatalf("expected invalid command error")
	}
	if err := router.On("ok", nil); err == nil {
		t.Fatalf("expected nil handler error")
	}
	if err := router.On("ok", func(context.Context, *client.Client, TextCommand) error { return nil }); err != nil {
		t.Fatalf("register command: %v", err)
	}
	if err := router.On("ok", func(context.Context, *client.Client, TextCommand) error { return nil }); err == nil {
		t.Fatalf("expected duplicate command error")
	}
}

func TestOnDirectCommandsIntegration(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	router := NewTextRouter()
	got := make(chan TextCommand, 1)
	if err := router.On("ping", func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		got <- cmd
		return nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}
	OnDirectCommands(rt, router)

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rt.Run(runCtx)

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[{"chatInfo":{"type":"direct","contact":{"contactId":5}},"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"/ping hi"}}}}]}}`)

	select {
	case cmd := <-got:
		if cmd.Name != "ping" || cmd.Args != "hi" || cmd.Message.ContactID != 5 {
			t.Fatalf("unexpected parsed command: %+v", cmd)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting direct command")
	}
}
