# Examples

All examples assume SimpleX CLI websocket API is running locally.

## 1. Start SimpleX CLI

```bash
simplex-chat -p 5225
```

## 2. Run one example

### Echo bot

```bash
go run ./examples/echo-bot
```

Try in chat:

- `/help`
- `/ping`
- `/echo hello`
- `/echo "hello world"`

### FAQ bot

```bash
go run ./examples/faq-bot
```

Try in chat:

- `/help`
- `/faq`
- `/faq pricing`
- send plain text: `what are your support hours?`

### Welcome bot

```bash
go run ./examples/welcome-bot
```

Try in chat:

- send any non-command message from a new contact (receives welcome)
- `/start`
- `/help`

### Moderation bot

```bash
go run ./examples/moderation
```

Try in chat:

- `/help`
- `/words`
- `/addword rude`
- send plain text containing blocked word
- `/delword rude`

### Original echo sample (advanced)

```bash
go run ./examples/echo
```

Uses the same core flow but with a slightly more verbose setup.
