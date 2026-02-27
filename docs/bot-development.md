# Bot Development Guide

This guide focuses on writing bots quickly while keeping code maintainable.

## Architecture layers

- `client.Client`: transport + request/response correlation + typed senders
- `bot.Runtime`: event loop and handler dispatch
- `bot.TextRouter`: command parsing for direct text messages

## Runtime patterns

### Typed event handlers

```go
rt, _ := bot.NewRuntime(cli)

bot.OnTyped(rt, types.EventTypeContactConnected,
    func(ctx context.Context, cli *client.Client, event types.EventContactConnected) error {
        return nil
    },
)
```

### Middleware

```go
rt.Use(func(next bot.Handler) bot.Handler {
    return func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
        started := time.Now()
        err := next(ctx, cli, msg)
        log.Printf("event=%s took=%s err=%v", msg.Resp.Type, time.Since(started), err)
        return err
    }
})
```

### Centralized error handling

```go
rt.OnError(func(ctx context.Context, err error) {
    log.Printf("runtime error: %v", err)
})
```

## TextRouter patterns

### Registering commands

```go
router := bot.NewTextRouter(
    bot.WithCommandPrefix("/"),
    bot.WithCommandRequirePrefix(true),
    bot.WithCommandCaseInsensitive(true),
)

_ = router.OnWithDescription("help", "show commands", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
    return cmd.Reply(ctx, cli, strings.Join(router.HelpLines(), "\n"))
})
```

### Parsing args

`cmd.Argv()` supports quotes and escapes.

Examples:

- `/echo hello world` -> `['hello', 'world']`
- `/echo "hello world"` -> `['hello world']`
- `/echo one\ two` -> `['one two']`

### Unknown and rate-limited hooks

```go
router.OnUnknown(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
    return cmd.Reply(ctx, cli, "unknown command")
})

_ = router.EnablePerContactRateLimit(20, time.Minute)
router.OnRateLimited(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
    return cmd.Reply(ctx, cli, "too many commands")
})
```

## Sending messages

### High-level helpers

```go
err := cli.SendTextToContact(ctx, contactID, "hello")
```

```go
err := cli.SendTextToGroupWithOptions(ctx, groupID, "hello", client.SendTextOptions{
    Live: true,
})
```

### Message updates/deletes/reactions

```go
_, _ = cli.UpdateTextMessageInContact(ctx, contactID, chatItemID, "new text", false)
_ = cli.DeleteChatItemsInContact(ctx, contactID, []int64{chatItemID}, client.CIDeleteModeInternal)
_ = cli.AddChatItemReaction(ctx, "@123", chatItemID, map[string]any{
    "type": "emoji",
    "emoji": ":thumbsup:",
})
```

## Reconnect supervisor

Use reconnect wrapper for long-running bots.

```go
err := bot.RunWebSocketWithReconnect(
    ctx,
    "wss://bot.example/ws",
    []ws.Option{ws.WithRequireWSS(true)},
    []client.Option{client.WithStrictResponses(false)},
    func(cli *client.Client) (bot.Runner, error) {
        rt, err := bot.NewRuntime(cli)
        if err != nil {
            return nil, err
        }
        bot.OnDirectCommands(rt, router)
        return rt, nil
    },
    bot.WithReconnectBackoff(1*time.Second, 30*time.Second),
    bot.WithReconnectStablePeriod(30*time.Second),
)
```

## Production coding conventions

- prefer typed handlers and typed sender results over raw JSON processing
- treat all external payloads as untrusted
- avoid global mutable state in handlers
- keep command handlers small and move business logic to services
- add table tests for command parsing and API response edge cases

## Official references

- Command semantics: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/COMMANDS.md
- Event payloads: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/EVENTS.md
- Type definitions: https://github.com/simplex-chat/simplex-chat/blob/stable/bots/api/TYPES.md

## Local runnable examples

- `go run ./examples/echo-bot`
- `go run ./examples/faq-bot`
- `go run ./examples/welcome-bot`
- `go run ./examples/moderation`

Quick scenario checklist is in `examples/README.md`.
