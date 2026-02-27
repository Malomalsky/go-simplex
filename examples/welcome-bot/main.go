package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

type seenContacts struct {
	mu   sync.Mutex
	seen map[int64]struct{}
}

func newSeenContacts() *seenContacts {
	return &seenContacts{seen: make(map[int64]struct{})}
}

func (s *seenContacts) FirstSeen(contactID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.seen[contactID]; exists {
		return false
	}
	s.seen[contactID] = struct{}{}
	return true
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seen := newSeenContacts()
	welcome := "Welcome. Commands: /help, /start"

	router := bot.NewTextRouter()
	if err := router.EnablePerContactRateLimit(30, time.Minute); err != nil {
		log.Fatalf("enable rate limit: %v", err)
	}

	if err := router.OnWithDescription("help", "show commands", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, strings.Join(router.HelpLines(), "\n"))
	}); err != nil {
		log.Fatalf("register help command: %v", err)
	}

	if err := router.OnWithDescription("start", "show welcome message", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, welcome)
	}); err != nil {
		log.Fatalf("register start command: %v", err)
	}

	router.OnUnknown(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "unknown command, try /help")
	})

	err := bot.RunWebSocketWithReconnect(
		ctx,
		"ws://localhost:5225",
		[]ws.Option{
			ws.WithReadLimit(16 << 20),
		},
		[]client.Option{client.WithStrictResponses(false)},
		func(cli *client.Client) (bot.Runner, error) {
			boot, err := cli.BootstrapBot(ctx)
			if err != nil {
				return nil, err
			}
			log.Printf("welcome bot address: %s", boot.Address)

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
				if seen.FirstSeen(msg.ContactID) {
					return msg.Reply(ctx, cli, welcome)
				}
				return nil
			})
			return rt, nil
		},
		bot.WithReconnectBackoff(1*time.Second, 20*time.Second),
		bot.WithReconnectMaxConsecutiveFailures(0),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
