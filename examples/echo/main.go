package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/command"
	"github.com/Malomalsky/go-simplex/sdk/protocol"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, err := client.NewWebSocket(ctx, "ws://localhost:5225")
	if err != nil {
		log.Fatalf("connect websocket: %v", err)
	}
	defer cli.Close(context.Background())

	boot, err := cli.BootstrapBot(ctx)
	if err != nil {
		log.Fatalf("bootstrap bot: %v", err)
	}

	fmt.Printf("Bot address: %s\n", boot.Address)

	rt, err := bot.NewRuntime(cli)
	if err != nil {
		log.Fatalf("new runtime: %v", err)
	}
	rt.OnError(func(ctx context.Context, err error) {
		log.Printf("runtime error: %v", err)
	})
	rt.On(string(types.EventTypeNewChatItems), func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
		messages, err := bot.ExtractDirectTextMessages(msg)
		if err != nil {
			return err
		}
		for _, m := range messages {
			reply := "echo: " + m.Text
			if err := cli.SendTextMessage(ctx, command.DirectRef(m.ContactID), reply); err != nil {
				return err
			}
			log.Printf("replied to contact %d", m.ContactID)
		}
		return nil
	})

	runCtx, runCancel := context.WithTimeout(ctx, 24*time.Hour)
	defer runCancel()

	if err := rt.Run(runCtx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatalf("runtime stopped with error: %v", err)
	}
}
