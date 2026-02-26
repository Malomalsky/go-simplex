package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Malomalsky/go-simplex/sdk/client"
)

type TextCommand struct {
	Name    string
	Args    string
	Message DirectTextMessage
}

type TextCommandHandler func(ctx context.Context, cli *client.Client, cmd TextCommand) error
type TextRouterOption func(*TextRouter)

type TextRouter struct {
	prefix          string
	requirePrefix   bool
	caseInsensitive bool

	commands map[string]TextCommandHandler
	unknown  TextCommandHandler
}

func NewTextRouter(opts ...TextRouterOption) *TextRouter {
	r := &TextRouter{
		prefix:        "/",
		requirePrefix: true,
		commands:      make(map[string]TextCommandHandler),
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

func (r *TextRouter) On(command string, h TextCommandHandler) error {
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
	r.commands[name] = h
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
	if h, exists := r.commands[name]; exists {
		return h(ctx, cli, cmd)
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
