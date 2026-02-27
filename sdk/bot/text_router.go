package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/Malomalsky/go-simplex/sdk/client"
)

type TextCommand struct {
	Name    string
	Args    string
	Message DirectTextMessage
}

func (c TextCommand) Reply(ctx context.Context, cli *client.Client, text string) error {
	return c.Message.Reply(ctx, cli, text)
}

func (c TextCommand) Argv() ([]string, error) {
	return parseCommandArgs(c.Args)
}

func (c TextCommand) Arg(index int) (string, bool) {
	argv, err := c.Argv()
	if err != nil || index < 0 || index >= len(argv) {
		return "", false
	}
	return argv[index], true
}

type TextCommandHandler func(ctx context.Context, cli *client.Client, cmd TextCommand) error
type TextRouterOption func(*TextRouter)

type textRoute struct {
	handler     TextCommandHandler
	description string
}

type TextRouter struct {
	prefix          string
	requirePrefix   bool
	caseInsensitive bool
	maxTextBytes    int

	commands map[string]textRoute
	unknown  TextCommandHandler
}

func NewTextRouter(opts ...TextRouterOption) *TextRouter {
	r := &TextRouter{
		prefix:        "/",
		requirePrefix: true,
		maxTextBytes:  4096,
		commands:      make(map[string]textRoute),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithCommandPrefix(prefix string) TextRouterOption {
	return func(r *TextRouter) {
		r.prefix = prefix
	}
}

func WithCommandRequirePrefix(require bool) TextRouterOption {
	return func(r *TextRouter) {
		r.requirePrefix = require
	}
}

func WithCommandCaseInsensitive(caseInsensitive bool) TextRouterOption {
	return func(r *TextRouter) {
		r.caseInsensitive = caseInsensitive
	}
}

func WithCommandMaxTextBytes(max int) TextRouterOption {
	return func(r *TextRouter) {
		if max >= 0 {
			r.maxTextBytes = max
		}
	}
}

func (r *TextRouter) On(command string, h TextCommandHandler) error {
	return r.OnWithDescription(command, "", h)
}

func (r *TextRouter) OnWithDescription(command string, description string, h TextCommandHandler) error {
	if h == nil {
		return fmt.Errorf("handler is nil")
	}
	name, ok := r.normalizeRegisteredCommand(command)
	if !ok {
		return fmt.Errorf("invalid command: %q", command)
	}
	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command already registered: %s", name)
	}
	r.commands[name] = textRoute{
		handler:     h,
		description: strings.TrimSpace(description),
	}
	return nil
}

func (r *TextRouter) OnUnknown(h TextCommandHandler) {
	r.unknown = h
}

func (r *TextRouter) Commands() []string {
	out := make([]string, 0, len(r.commands))
	for name := range r.commands {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (r *TextRouter) HelpLines() []string {
	names := r.Commands()
	if len(names) == 0 {
		return nil
	}
	lines := make([]string, 0, len(names))
	for _, name := range names {
		route := r.commands[name]
		label := name
		if r.prefix != "" {
			label = r.prefix + name
		}
		if route.description == "" {
			lines = append(lines, label)
			continue
		}
		lines = append(lines, label+" - "+route.description)
	}
	return lines
}

func (r *TextRouter) Handle(ctx context.Context, cli *client.Client, msg DirectTextMessage) error {
	name, args, ok := r.parse(msg.Text)
	if !ok {
		return nil
	}
	cmd := TextCommand{
		Name:    name,
		Args:    args,
		Message: msg,
	}
	if route, exists := r.commands[name]; exists {
		return route.handler(ctx, cli, cmd)
	}
	if r.unknown != nil {
		return r.unknown(ctx, cli, cmd)
	}
	return nil
}

func OnDirectCommands(rt *Runtime, router *TextRouter) {
	if rt == nil || router == nil {
		return
	}
	OnDirectText(rt, func(ctx context.Context, cli *client.Client, msg DirectTextMessage) error {
		return router.Handle(ctx, cli, msg)
	})
}

func (r *TextRouter) normalizeRegisteredCommand(command string) (string, bool) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", false
	}
	if strings.ContainsAny(command, " \t\r\n") {
		return "", false
	}
	if r.prefix != "" && strings.HasPrefix(command, r.prefix) {
		command = strings.TrimPrefix(command, r.prefix)
	}
	if command == "" {
		return "", false
	}
	if r.caseInsensitive {
		command = strings.ToLower(command)
	}
	return command, true
}

func (r *TextRouter) parse(text string) (name, args string, ok bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", false
	}
	if r.maxTextBytes > 0 && len(text) > r.maxTextBytes {
		return "", "", false
	}

	cmdToken := text
	if i := strings.IndexAny(text, " \t\r\n"); i >= 0 {
		cmdToken = text[:i]
		args = strings.TrimSpace(text[i+1:])
	}

	if r.prefix != "" && strings.HasPrefix(cmdToken, r.prefix) {
		cmdToken = strings.TrimPrefix(cmdToken, r.prefix)
	} else if r.requirePrefix {
		return "", "", false
	}

	if cmdToken == "" || strings.ContainsAny(cmdToken, " \t\r\n") {
		return "", "", false
	}
	if r.caseInsensitive {
		cmdToken = strings.ToLower(cmdToken)
	}
	return cmdToken, args, true
}

func parseCommandArgs(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	var (
		args         []string
		buf          strings.Builder
		quote        rune
		escape       bool
		tokenStarted bool
	)

	flush := func() {
		if !tokenStarted {
			return
		}
		args = append(args, buf.String())
		buf.Reset()
		tokenStarted = false
	}

	for _, r := range s {
		if escape {
			tokenStarted = true
			buf.WriteRune(r)
			escape = false
			continue
		}
		if r == '\\' {
			tokenStarted = true
			escape = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			tokenStarted = true
			buf.WriteRune(r)
			continue
		}
		if r == '"' || r == '\'' {
			tokenStarted = true
			quote = r
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		tokenStarted = true
		buf.WriteRune(r)
	}

	if escape {
		return nil, fmt.Errorf("unterminated escape sequence")
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quoted argument")
	}

	flush()
	return args, nil
}
