package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/protocol"
)

var errMockClosed = errors.New("mock transport closed")

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
		return nil, errMockClosed
	case payload := <-m.readCh:
		return payload, nil
	}
}

func (m *mockTransport) Write(ctx context.Context, payload []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.closed:
		return errMockClosed
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

func TestSendRawRoutesCorrelatedResponse(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type sendResult struct {
		msg protocol.Message
		err error
	}
	resultCh := make(chan sendResult, 1)
	go func() {
		msg, sendErr := c.SendRaw(ctx, "/user")
		resultCh <- sendResult{msg: msg, err: sendErr}
	}()

	rawReq := <-transport.writeCh
	var req protocol.CommandRequest
	if err := json.Unmarshal(rawReq, &req); err != nil {
		t.Fatalf("decode sent request: %v", err)
	}
	if req.CorrID == "" {
		t.Fatalf("corrId should not be empty")
	}

	transport.readCh <- []byte(fmt.Sprintf(`{"corrId":"%s","resp":{"type":"activeUser","user":{"userId":1}}}`, req.CorrID))

	res := <-resultCh
	if res.err != nil {
		t.Fatalf("send result error: %v", res.err)
	}
	if got, want := res.msg.CorrID, req.CorrID; got != want {
		t.Fatalf("corrId mismatch: got %q want %q", got, want)
	}
	if got, want := res.msg.Resp.Type, "activeUser"; got != want {
		t.Fatalf("response type mismatch: got %q want %q", got, want)
	}
}

func TestEventsAreDelivered(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport, WithEventBuffer(1))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	transport.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)

	select {
	case evt := <-c.Events():
		if !evt.IsEvent() {
			t.Fatalf("expected event frame")
		}
		if got, want := evt.Resp.Type, "newChatItems"; got != want {
			t.Fatalf("event type mismatch: got %q want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for event")
	}
}

func TestUnknownCorrIDReportsError(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport, WithErrorBuffer(1))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	transport.readCh <- []byte(`{"corrId":"999","resp":{"type":"cmdOk"}}`)

	select {
	case err := <-c.Errors():
		if err == nil {
			t.Fatalf("expected non-nil error")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for error event")
	}
}

