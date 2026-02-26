package bot

import (
	"context"
	"fmt"
	"sync"

	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/protocol"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

type Handler func(ctx context.Context, cli *client.Client, msg protocol.Message) error
type ErrorHandler func(ctx context.Context, err error)

type Runtime struct {
	client *client.Client

	mu       sync.RWMutex
	handlers map[string][]Handler
	any      []Handler

	onError ErrorHandler
}

func NewRuntime(cli *client.Client) (*Runtime, error) {
	if cli == nil {
		return nil, fmt.Errorf("client is nil")
	}
	return &Runtime{
		client:   cli,
		handlers: make(map[string][]Handler),
		onError:  func(context.Context, error) {},
	}, nil
}

func (r *Runtime) On(eventType string, h Handler) {
	if eventType == "" || h == nil {
		return
	}
	r.mu.Lock()
	r.handlers[eventType] = append(r.handlers[eventType], h)
	r.mu.Unlock()
}

func (r *Runtime) OnEvent(eventType types.EventType, h Handler) {
	r.On(string(eventType), h)
}

func (r *Runtime) OnAny(h Handler) {
	if h == nil {
		return
	}
	r.mu.Lock()
	r.any = append(r.any, h)
	r.mu.Unlock()
}

func (r *Runtime) OnError(h ErrorHandler) {
	if h == nil {
		return
	}
	r.mu.Lock()
	r.onError = h
	r.mu.Unlock()
}

func (r *Runtime) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-r.client.Errors():
			if !ok {
				return nil
			}
			r.emitErr(ctx, err)
		case msg, ok := <-r.client.Events():
			if !ok {
				return nil
			}
			if err := r.dispatch(ctx, msg); err != nil {
				r.emitErr(ctx, err)
			}
		}
	}
}

func (r *Runtime) dispatch(ctx context.Context, msg protocol.Message) error {
	r.mu.RLock()
	specific := append([]Handler(nil), r.handlers[msg.Resp.Type]...)
	any := append([]Handler(nil), r.any...)
	r.mu.RUnlock()

	for _, h := range specific {
		if err := h(ctx, r.client, msg); err != nil {
			return fmt.Errorf("event %s handler error: %w", msg.Resp.Type, err)
		}
	}
	for _, h := range any {
		if err := h(ctx, r.client, msg); err != nil {
			return fmt.Errorf("event %s any-handler error: %w", msg.Resp.Type, err)
		}
	}
	return nil
}

func (r *Runtime) emitErr(ctx context.Context, err error) {
	if err == nil {
		return
	}
	r.mu.RLock()
	handler := r.onError
	r.mu.RUnlock()
	handler(ctx, err)
}
