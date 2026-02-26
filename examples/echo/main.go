package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/bot"
	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/protocol"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, err := client.NewWebSocket(ctx, "ws://localhost:5225")
	if err != nil {
		log.Fatalf("connect websocket: %v", err)
	}
	defer cli.Close(context.Background())

	user, err := cli.GetActiveUser(ctx)
	if err != nil {
		log.Fatalf("get active user: %v", err)
	}

	address, err := cli.EnsureUserAddress(ctx, user.UserID)
	if err != nil {
		log.Fatalf("ensure user address: %v", err)
	}
	if err := cli.EnableAddressAutoAccept(ctx, user.UserID); err != nil {
		log.Fatalf("enable auto-accept: %v", err)
	}

	fmt.Printf("Bot address: %s\n", address)

	rt, err := bot.NewRuntime(cli)
	if err != nil {
		log.Fatalf("new runtime: %v", err)
	}
	rt.OnError(func(ctx context.Context, err error) {
		log.Printf("runtime error: %v", err)
	})
	rt.On("newChatItems", func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
		// For now we keep echo logic minimal; full typed chat item decoding will be added
		// with generated types in the next iterations.
		log.Printf("received newChatItems event")
		return nil
	})

	runCtx, runCancel := context.WithTimeout(ctx, 24*time.Hour)
	defer runCancel()

	if err := rt.Run(runCtx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatalf("runtime stopped with error: %v", err)
	}
}
