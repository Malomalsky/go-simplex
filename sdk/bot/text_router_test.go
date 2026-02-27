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

func TestTextCommandArgv(t *testing.T) {
	t.Parallel()

	cmd := TextCommand{
		Name: "echo",
		Args: `one "two words" 'three words' four\ five`,
	}
	argv, err := cmd.Argv()
	if err != nil {
		t.Fatalf("parse argv: %v", err)
	}
	if len(argv) != 4 {
		t.Fatalf("unexpected argv size: %d (%#v)", len(argv), argv)
	}
	if argv[0] != "one" || argv[1] != "two words" || argv[2] != "three words" || argv[3] != "four five" {
		t.Fatalf("unexpected argv: %#v", argv)
	}

	if got, ok := cmd.Arg(1); !ok || got != "two words" {
		t.Fatalf("unexpected arg(1): %q %v", got, ok)
	}
	if _, ok := cmd.Arg(8); ok {
		t.Fatalf("expected missing arg for out-of-range index")
	}
}

func TestTextCommandArgvKeepsEmptyQuotedArg(t *testing.T) {
	t.Parallel()

	cmd := TextCommand{
		Name: "echo",
		Args: `"" "a"`,
	}
	argv, err := cmd.Argv()
	if err != nil {
		t.Fatalf("parse argv: %v", err)
	}
	if len(argv) != 2 {
		t.Fatalf("unexpected argv size: %d (%#v)", len(argv), argv)
	}
	if argv[0] != "" || argv[1] != "a" {
		t.Fatalf("unexpected argv values: %#v", argv)
	}
}

func TestTextCommandArgvErrors(t *testing.T) {
	t.Parallel()

	cmd := TextCommand{Name: "broken", Args: `"unterminated`}
	if _, err := cmd.Argv(); err == nil {
		t.Fatalf("expected argv parsing error")
	}
}

func TestTextCommandReply(t *testing.T) {
	t.Parallel()

	cmd := TextCommand{
		Name: "echo",
		Message: DirectTextMessage{
			ContactID: 1,
		},
	}
	if err := cmd.Reply(context.Background(), nil, "x"); err == nil {
		t.Fatalf("expected nil client error")
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

func TestTextRouterOnWithDescriptionAndHelpLines(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	if err := router.OnWithDescription("help", "show available commands", func(context.Context, *client.Client, TextCommand) error {
		return nil
	}); err != nil {
		t.Fatalf("register help command: %v", err)
	}
	if err := router.OnWithDescription("ping", "check bot liveness", func(context.Context, *client.Client, TextCommand) error {
		return nil
	}); err != nil {
		t.Fatalf("register ping command: %v", err)
	}

	lines := router.HelpLines()
	if len(lines) != 2 {
		t.Fatalf("unexpected help line count: %d", len(lines))
	}
	if lines[0] != "/help - show available commands" {
		t.Fatalf("unexpected first help line: %q", lines[0])
	}
	if lines[1] != "/ping - check bot liveness" {
		t.Fatalf("unexpected second help line: %q", lines[1])
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

func TestTextRouterMaxTextBytes(t *testing.T) {
	t.Parallel()

	router := NewTextRouter(WithCommandMaxTextBytes(5))
	called := false
	if err := router.On("ping", func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}

	if err := router.Handle(context.Background(), nil, DirectTextMessage{Text: "/ping long"}); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if called {
		t.Fatalf("expected command over max bytes to be ignored")
	}
}

func TestTextRouterEnablePerContactRateLimit(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	if err := router.EnablePerContactRateLimit(2, time.Minute); err != nil {
		t.Fatalf("enable rate limit: %v", err)
	}

	hits := 0
	if err := router.On("ping", func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		hits++
		return nil
	}); err != nil {
		t.Fatalf("register command: %v", err)
	}

	limited := 0
	router.OnRateLimited(func(ctx context.Context, cli *client.Client, cmd TextCommand) error {
		limited++
		return nil
	})

	msg := DirectTextMessage{ContactID: 7, Text: "/ping"}
	_ = router.Handle(context.Background(), nil, msg)
	_ = router.Handle(context.Background(), nil, msg)
	_ = router.Handle(context.Background(), nil, msg)

	if hits != 2 {
		t.Fatalf("unexpected handled count: %d", hits)
	}
	if limited != 1 {
		t.Fatalf("unexpected limited count: %d", limited)
	}
}

func TestTextRouterEnablePerContactRateLimitValidation(t *testing.T) {
	t.Parallel()

	router := NewTextRouter()
	if err := router.EnablePerContactRateLimit(0, time.Minute); err == nil {
		t.Fatalf("expected invalid rate-limit config error")
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
