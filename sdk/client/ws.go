package client

import (
	"context"

	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

func NewWebSocket(ctx context.Context, url string, clientOptions ...Option) (*Client, error) {
	return NewWebSocketWithOptions(ctx, url, nil, clientOptions...)
}

func NewWebSocketWithOptions(ctx context.Context, url string, wsOptions []ws.Option, clientOptions ...Option) (*Client, error) {
	transport, err := ws.Dial(ctx, url, wsOptions...)
	if err != nil {
		return nil, err
	}
	return New(transport, clientOptions...)
}
