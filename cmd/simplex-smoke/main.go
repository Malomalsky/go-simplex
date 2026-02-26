package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
)

func main() {
	var wsURL string
	flag.StringVar(&wsURL, "ws", "ws://localhost:5225", "simplex websocket url")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cli, err := client.NewWebSocket(ctx, wsURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer cli.Close(context.Background())

	user, err := cli.GetActiveUser(ctx)
	if err != nil {
		log.Fatalf("get active user: %v", err)
	}

	fmt.Printf("ok: user_id=%d display_name=%s\n", user.UserID, user.Profile.DisplayName)
}
