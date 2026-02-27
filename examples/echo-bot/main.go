package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := bot.NewTextRouter()
	if err := router.EnablePerContactRateLimit(30, time.Minute); err != nil {
		log.Fatalf("enable rate limit: %v", err)
	}

	if err := router.OnWithDescription("help", "show commands", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, strings.Join(router.HelpLines(), "\n"))
	}); err != nil {
		log.Fatalf("register help command: %v", err)
	}

	if err := router.OnWithDescription("ping", "health check", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "pong")
	}); err != nil {
		log.Fatalf("register ping command: %v", err)
	}

	if err := router.OnWithDescription("echo", "echo text back", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		argv, err := cmd.Argv()
		if err != nil {
			return cmd.Reply(ctx, cli, "invalid args: "+err.Error())
		}
		if len(argv) == 0 {
			return cmd.Reply(ctx, cli, "usage: /echo <text>")
		}
		return cmd.Reply(ctx, cli, strings.Join(argv, " "))
	}); err != nil {
		log.Fatalf("register echo command: %v", err)
	}

	router.OnUnknown(func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Reply(ctx, cli, "unknown command, try /help")
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
			log.Printf("echo bot address: %s", boot.Address)

			rt, err := bot.NewRuntime(cli)
			if err != nil {
				return nil, err
			}
			rt.OnError(func(ctx context.Context, err error) {
				log.Printf("runtime error: %v", err)
			})
			bot.OnDirectCommands(rt, router)
			return rt, nil
		},
		bot.WithReconnectBackoff(1*time.Second, 20*time.Second),
		bot.WithReconnectMaxConsecutiveFailures(0),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
