package main

import (
	"context"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

type denyList struct {
	mu    sync.RWMutex
	words map[string]struct{}
}

func newDenyList(initial []string) *denyList {
	d := &denyList{words: make(map[string]struct{}, len(initial))}
	for _, word := range initial {
		d.Add(word)
	}
	return d
}

func (d *denyList) Add(word string) {
	word = normalizeWord(word)
	if word == "" {
		return
	}
	d.mu.Lock()
	d.words[word] = struct{}{}
	d.mu.Unlock()
}

func (d *denyList) Remove(word string) bool {
	word = normalizeWord(word)
	if word == "" {
		return false
	}
	d.mu.Lock()
	_, exists := d.words[word]
	if exists {
		delete(d.words, word)
	}
	d.mu.Unlock()
	return exists
}

func (d *denyList) List() []string {
	d.mu.RLock()
	out := make([]string, 0, len(d.words))
	for word := range d.words {
		out = append(out, word)
	}
	d.mu.RUnlock()
	sort.Strings(out)
	return out
}

func (d *denyList) Match(text string) (string, bool) {
	normalized := strings.ToLower(text)
	d.mu.RLock()
	defer d.mu.RUnlock()
	for word := range d.words {
		if strings.Contains(normalized, word) {
			return word, true
		}
	}
	return "", false
}

func normalizeWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	list := newDenyList([]string{"spam", "scam"})

	router := bot.NewTextRouter()
	if err := router.EnablePerContactRateLimit(20, time.Minute); err != nil {
		log.Fatalf("enable rate limit: %v", err)
	}
	if err := router.OnWithDescription("help", "show command list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, strings.Join(router.HelpLines(), "\n"))
	}); err != nil {
		log.Fatalf("register help command: %v", err)
	}
	if err := router.OnWithDescription("addword", "add word to deny-list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		word, ok := cmd.Arg(0)
		if !ok {
			return cmd.Reply(ctx, cli, "usage: /addword <word>")
		}
		list.Add(word)
		return cmd.Reply(ctx, cli, "added: "+normalizeWord(word))
	}); err != nil {
		log.Fatalf("register addword command: %v", err)
	}
	if err := router.OnWithDescription("delword", "remove word from deny-list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		word, ok := cmd.Arg(0)
		if !ok {
			return cmd.Reply(ctx, cli, "usage: /delword <word>")
		}
		if !list.Remove(word) {
			return cmd.Reply(ctx, cli, "not found: "+normalizeWord(word))
		}
		return cmd.Reply(ctx, cli, "removed: "+normalizeWord(word))
	}); err != nil {
		log.Fatalf("register delword command: %v", err)
	}
	if err := router.OnWithDescription("words", "show deny-list", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		words := list.List()
		if len(words) == 0 {
			return cmd.Reply(ctx, cli, "deny-list is empty")
		}
		return cmd.Reply(ctx, cli, "deny-list: "+strings.Join(words, ", "))
	}); err != nil {
		log.Fatalf("register words command: %v", err)
	}
	if err := router.OnWithDescription("ping", "health check", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "pong")
	}); err != nil {
		log.Fatalf("register ping command: %v", err)
	}

	router.OnUnknown(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "unknown command. try /help")
	})
	router.OnRateLimited(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "rate limit exceeded, try again later")
	})

	err := bot.RunWebSocketWithReconnect(
		ctx,
		"ws://localhost:5225",
		[]ws.Option{
			// For remote deployments use: ws.WithRequireWSS(true).
			ws.WithReadLimit(16 << 20),
		},
		[]client.Option{client.WithStrictResponses(false)},
		func(cli *client.Client) (bot.Runner, error) {
			boot, err := cli.BootstrapBot(ctx)
			if err != nil {
				return nil, err
			}
			log.Printf("bot address: %s", boot.Address)

			rt, err := bot.NewRuntime(cli)
			if err != nil {
				return nil, err
			}
			rt.OnError(func(ctx context.Context, err error) {
				log.Printf("runtime error: %v", err)
			})
			bot.OnDirectCommands(rt, router)
			bot.OnDirectText(rt, func(ctx context.Context, cli *client.Client, msg bot.DirectTextMessage) error {
				if strings.HasPrefix(strings.TrimSpace(msg.Text), "/") {
					return nil
				}
				if word, matched := list.Match(msg.Text); matched {
					return msg.Reply(ctx, cli, "message blocked by moderation rule: "+word)
				}
				return nil
			})
			return rt, nil
		},
		bot.WithReconnectBackoff(1*time.Second, 20*time.Second),
		bot.WithReconnectMaxConsecutiveFailures(0),
		bot.WithReconnectErrorHandler(func(err error) {
			log.Printf("reconnect: %v", err)
		}),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
