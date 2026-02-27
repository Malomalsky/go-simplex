package main

import (
	"context"
	"errors"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	answers := map[string]string{
		"pricing": "Pricing: /faq pricing -> Starter $9, Team $29, Enterprise custom",
		"support": "Support: support@example.com, weekdays 09:00-18:00 UTC",
		"hours":   "Working hours: Mon-Fri, 09:00-18:00 UTC",
	}
	topics := sortedKeys(answers)

	router := bot.NewTextRouter()
	if err := router.EnablePerContactRateLimit(30, time.Minute); err != nil {
		log.Fatalf("enable rate limit: %v", err)
	}

	if err := router.OnWithDescription("help", "show commands", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, strings.Join(router.HelpLines(), "\n"))
	}); err != nil {
		log.Fatalf("register help command: %v", err)
	}

	if err := router.OnWithDescription("faq", "show FAQ topics or /faq <topic>", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		topic, ok := cmd.Arg(0)
		if !ok {
			return cmd.Reply(ctx, cli, "topics: "+strings.Join(topics, ", ")+"\nusage: /faq <topic>")
		}
		topic = strings.ToLower(strings.TrimSpace(topic))
		answer, exists := answers[topic]
		if !exists {
			return cmd.Reply(ctx, cli, "unknown topic: "+topic+"\navailable: "+strings.Join(topics, ", "))
		}
		return cmd.Reply(ctx, cli, answer)
	}); err != nil {
		log.Fatalf("register faq command: %v", err)
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
			log.Printf("faq bot address: %s", boot.Address)

			rt, err := bot.NewRuntime(cli)
			if err != nil {
				return nil, err
			}
			rt.OnError(func(ctx context.Context, err error) {
				log.Printf("runtime error: %v", err)
			})

			bot.OnDirectCommands(rt, router)
			bot.OnDirectText(rt, func(ctx context.Context, cli *client.Client, msg bot.DirectTextMessage) error {
				text := strings.ToLower(strings.TrimSpace(msg.Text))
				if strings.HasPrefix(text, "/") {
					return nil
				}
				switch {
				case strings.Contains(text, "price") || strings.Contains(text, "pricing"):
					return msg.Reply(ctx, cli, "Looks like pricing question. Try: /faq pricing")
				case strings.Contains(text, "support") || strings.Contains(text, "help"):
					return msg.Reply(ctx, cli, "Need support details? Try: /faq support")
				case strings.Contains(text, "hours") || strings.Contains(text, "time"):
					return msg.Reply(ctx, cli, "Working hours are in: /faq hours")
				default:
					return nil
				}
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

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
