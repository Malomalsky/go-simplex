package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := bot.NewTextRouter()
	if err := router.On("ping", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		return cmd.Message.Reply(ctx, cli, "pong")
	}); err != nil {
		log.Fatalf("register ping command: %v", err)
	}
	if err := router.On("echo", func(ctx context.Context, cli *client.Client, cmd bot.TextCommand) error {
		reply := cmd.Args
		if reply == "" {
			reply = cmd.Message.Text
		}
		if err := cmd.Message.Reply(ctx, cli, "echo: "+reply); err != nil {
			return err
		}
		log.Printf("replied to contact %d", cmd.Message.ContactID)
		return nil
	}); err != nil {
		log.Fatalf("register echo command: %v", err)
	}

	err := bot.RunWebSocketWithReconnect(
		ctx,
		"ws://localhost:5225",
		[]ws.Option{
			// For remote deployments use: ws.WithRequireWSS(true).
			ws.WithReadLimit(16 << 20),
		},
		[]client.Option{
			client.WithStrictResponses(false),
		},
		func(cli *client.Client) (bot.Runner, error) {
			boot, err := cli.BootstrapBot(ctx)
			if err != nil {
				return nil, fmt.Errorf("bootstrap bot: %w", err)
			}
			fmt.Printf("Bot address: %s\n", boot.Address)

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
		bot.WithReconnectMaxConsecutiveFailures(0), // unlimited
		bot.WithReconnectErrorHandler(func(err error) {
			log.Printf("reconnect: %v", err)
		}),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
