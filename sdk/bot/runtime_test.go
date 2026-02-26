package bot

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/protocol"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

var errTestClosed = errors.New("test transport closed")

type mockTransport struct {
	readCh  chan []byte
	writeCh chan []byte
	closed  chan struct{}
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		readCh:  make(chan []byte, 8),
		writeCh: make(chan []byte, 8),
		closed:  make(chan struct{}),
	}
}

func (m *mockTransport) Read(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-m.closed:
		return nil, errTestClosed
	case payload := <-m.readCh:
		return payload, nil
	}
}

func (m *mockTransport) Write(ctx context.Context, payload []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.closed:
		return errTestClosed
	case m.writeCh <- payload:
		return nil
	}
}

func (m *mockTransport) Close() error {
	select {
	case <-m.closed:
	default:
		close(m.closed)
	}
	return nil
}

func TestRuntimeDispatch(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	var mu sync.Mutex
	var called []string

	rt.On("newChatItems", func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
		mu.Lock()
		called = append(called, "specific")
		mu.Unlock()
		return nil
	})
	rt.OnAny(func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
		mu.Lock()
		called = append(called, "any")
		mu.Unlock()
		return nil
	})
	rt.OnEvent(types.EventTypeContactUpdated, func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
		return nil
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- rt.Run(runCtx)
	}()

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for handlers")
		default:
			mu.Lock()
			ok := len(called) == 2
			mu.Unlock()
			if ok {
				cancel()
				<-done
				mu.Lock()
				defer mu.Unlock()
				if called[0] != "specific" || called[1] != "any" {
					t.Fatalf("unexpected handler order: %#v", called)
				}
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRuntimeOnError(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	errCh := make(chan error, 1)
	rt.On("newChatItems", func(context.Context, *client.Client, protocol.Message) error {
		return errors.New("boom")
	})
	rt.OnError(func(ctx context.Context, err error) {
		errCh <- err
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rt.Run(runCtx)

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatalf("expected non-nil handler error")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for runtime error")
	}
}

func TestRuntimeOnTyped(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	got := make(chan types.EventNewChatItems, 1)
	OnTyped(rt, types.EventTypeNewChatItems, func(ctx context.Context, cli *client.Client, event types.EventNewChatItems) error {
		got <- event
		return nil
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rt.Run(runCtx)

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","user":{"userId":1,"profile":{"displayName":"bot"}},"chatItems":[]}}`)

	select {
	case evt := <-got:
		if evt.Type != types.EventTypeNewChatItems {
			t.Fatalf("unexpected event type: %s", evt.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting typed handler")
	}
}

func TestRuntimeOnDirectText(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	got := make(chan DirectTextMessage, 1)
	OnDirectText(rt, func(ctx context.Context, cli *client.Client, msg DirectTextMessage) error {
		got <- msg
		return nil
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rt.Run(runCtx)

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[{"chatInfo":{"type":"direct","contact":{"contactId":9}},"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"ping"}}}},{"chatInfo":{"type":"group"},"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"ignore"}}}}]}}`)

	select {
	case msg := <-got:
		if msg.ContactID != 9 || msg.Text != "ping" {
			t.Fatalf("unexpected direct text message: %+v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting direct text handler")
	}
}

func TestRuntimeMiddlewareOrder(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	var mu sync.Mutex
	var called []string
	push := func(s string) {
		mu.Lock()
		called = append(called, s)
		mu.Unlock()
	}

	rt.Use(func(next Handler) Handler {
		return func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
			push("mw1-pre")
			err := next(ctx, cli, msg)
			push("mw1-post")
			return err
		}
	})
	rt.Use(func(next Handler) Handler {
		return func(ctx context.Context, cli *client.Client, msg protocol.Message) error {
			push("mw2-pre")
			err := next(ctx, cli, msg)
			push("mw2-post")
			return err
		}
	})
	rt.On("newChatItems", func(context.Context, *client.Client, protocol.Message) error {
		push("handler")
		return nil
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- rt.Run(runCtx)
	}()

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for middleware order")
		default:
			mu.Lock()
			ok := len(called) == 5
			got := append([]string(nil), called...)
			mu.Unlock()
			if ok {
				cancel()
				<-done
				want := []string{"mw1-pre", "mw2-pre", "handler", "mw2-post", "mw1-post"}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("unexpected middleware order: %#v", got)
					}
				}
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRuntimeHandlerPanicRecovery(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	cli, err := client.New(tr)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer cli.Close(context.Background())

	rt, err := NewRuntime(cli)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	rt.On("newChatItems", func(context.Context, *client.Client, protocol.Message) error {
		panic("boom")
	})

	errCh := make(chan error, 1)
	rt.OnError(func(ctx context.Context, err error) {
		errCh <- err
	})

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rt.Run(runCtx)

	tr.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatalf("expected non-nil runtime error")
		}
		if !strings.Contains(err.Error(), "panic in handler") {
			t.Fatalf("unexpected panic error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting panic error")
	}
}
