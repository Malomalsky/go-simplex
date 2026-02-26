package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/command"
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

func TestSendUsesTypedRequest(t *testing.T) {
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
		msg, sendErr := c.Send(ctx, command.ShowActiveUser{})
		resultCh <- sendResult{msg: msg, err: sendErr}
	}()

	rawReq := <-transport.writeCh
	var req protocol.CommandRequest
	if err := json.Unmarshal(rawReq, &req); err != nil {
		t.Fatalf("decode sent request: %v", err)
	}
	if req.Cmd != "/user" {
		t.Fatalf("unexpected command string: %q", req.Cmd)
	}

	transport.readCh <- []byte(fmt.Sprintf(`{"corrId":"%s","resp":{"type":"activeUser","user":{"userId":1}}}`, req.CorrID))
	res := <-resultCh
	if res.err != nil {
		t.Fatalf("send typed result error: %v", res.err)
	}
}

func TestSendRejectsUnexpectedResponseType(t *testing.T) {
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
		msg, sendErr := c.Send(ctx, command.ShowActiveUser{})
		resultCh <- sendResult{msg: msg, err: sendErr}
	}()

	rawReq := <-transport.writeCh
	var req protocol.CommandRequest
	if err := json.Unmarshal(rawReq, &req); err != nil {
		t.Fatalf("decode sent request: %v", err)
	}
	transport.readCh <- []byte(fmt.Sprintf(`{"corrId":"%s","resp":{"type":"cmdOk"}}`, req.CorrID))

	res := <-resultCh
	var typedErr *UnexpectedResponseTypeError
	if !errors.As(res.err, &typedErr) {
		t.Fatalf("expected UnexpectedResponseTypeError, got: %v", res.err)
	}
	if typedErr.ResponseType != "cmdOk" {
		t.Fatalf("unexpected response type: %q", typedErr.ResponseType)
	}
}

func TestGeneratedSenderMethod(t *testing.T) {
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
		res APICreateMyAddressResult
		err error
	}
	resultCh := make(chan sendResult, 1)
	go func() {
		res, sendErr := c.SendAPICreateMyAddress(ctx, command.APICreateMyAddress{UserId: 1})
		resultCh <- sendResult{res: res, err: sendErr}
	}()

	rawReq := <-transport.writeCh
	var req protocol.CommandRequest
	if err := json.Unmarshal(rawReq, &req); err != nil {
		t.Fatalf("decode sent request: %v", err)
	}
	transport.readCh <- []byte(fmt.Sprintf(`{"corrId":"%s","resp":{"type":"userContactLinkCreated","user":{"userId":1},"connLinkContact":{"connFullLink":"smp://full","connShortLink":"smp://short"}}}`, req.CorrID))

	res := <-resultCh
	if res.err != nil {
		t.Fatalf("generated sender error: %v", res.err)
	}
	if res.res.UserContactLinkCreated == nil {
		t.Fatalf("expected UserContactLinkCreated response")
	}
	if res.res.ChatCmdError != nil {
		t.Fatalf("unexpected ChatCmdError response")
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

type panicRequest struct{}

func (panicRequest) CommandString() string {
	panic("boom")
}

func TestSendRecoversCommandStringPanic(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	_, err = c.Send(context.Background(), panicRequest{})
	if err == nil {
		t.Fatalf("expected error for command string panic")
	}
	if !strings.Contains(err.Error(), "build command string panic") {
		t.Fatalf("unexpected panic conversion error: %v", err)
	}
}

func TestSendRawRejectsByValidator(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport, WithRawCommandValidator(func(cmd string) error {
		if strings.HasPrefix(cmd, "/user") {
			return nil
		}
		return fmt.Errorf("blocked by validator")
	}))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	_, err = c.SendRaw(context.Background(), "/_delete @1 entity")
	if err == nil {
		t.Fatalf("expected validator rejection")
	}
	if !strings.Contains(err.Error(), "raw command rejected") {
		t.Fatalf("unexpected validator error: %v", err)
	}

	select {
	case <-transport.writeCh:
		t.Fatalf("validator-rejected command should not be written")
	default:
	}
}

func TestSendRawAllowPrefixes(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport, WithRawCommandAllowPrefixes("/user", "/_groups "))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resultCh := make(chan error, 1)
	go func() {
		_, sendErr := c.SendRaw(ctx, "/user")
		resultCh <- sendErr
	}()

	rawReq := <-transport.writeCh
	var req protocol.CommandRequest
	if err := json.Unmarshal(rawReq, &req); err != nil {
		t.Fatalf("decode sent request: %v", err)
	}
	transport.readCh <- []byte(fmt.Sprintf(`{"corrId":"%s","resp":{"type":"cmdOk"}}`, req.CorrID))

	if err := <-resultCh; err != nil {
		t.Fatalf("allowed raw command should succeed: %v", err)
	}

	if _, err := c.SendRaw(context.Background(), "/_delete @1 entity"); err == nil {
		t.Fatalf("expected disallowed command to be rejected")
	}
}

func TestSendRawRejectsControlCharacters(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	_, err = c.SendRaw(context.Background(), "/user\n/_delete @1 entity")
	if err == nil {
		t.Fatalf("expected control character rejection")
	}
	if !strings.Contains(err.Error(), "control character") {
		t.Fatalf("unexpected control char error: %v", err)
	}

	select {
	case <-transport.writeCh:
		t.Fatalf("invalid command should not be written")
	default:
	}
}

func TestSendRawRejectsTooLargeCommand(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport, WithRawCommandMaxBytes(4))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	_, err = c.SendRaw(context.Background(), "/user")
	if err == nil {
		t.Fatalf("expected max size rejection")
	}
	if !strings.Contains(err.Error(), "exceeds max size") {
		t.Fatalf("unexpected max size error: %v", err)
	}
}

func TestSendRawRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	invalid := string([]byte{'/', 0xff})
	_, err = c.SendRaw(context.Background(), invalid)
	if err == nil {
		t.Fatalf("expected invalid UTF-8 rejection")
	}
	if !strings.Contains(err.Error(), "invalid UTF-8") {
		t.Fatalf("unexpected UTF-8 error: %v", err)
	}
}

func TestEventOverflowDropNewest(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var dropped []string

	transport := newMockTransport()
	c, err := New(
		transport,
		WithEventBuffer(1),
		WithEventOverflowPolicy(OverflowPolicyDropNewest),
		WithDropHandler(func(kind string, n uint64) {
			mu.Lock()
			dropped = append(dropped, fmt.Sprintf("%s:%d", kind, n))
			mu.Unlock()
		}),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	transport.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)
	transport.readCh <- []byte(`{"resp":{"type":"newChatItems","chatItems":[]}}`)

	deadline := time.After(2 * time.Second)
	for {
		if c.DroppedEvents() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting dropped event counter")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	select {
	case <-c.Events():
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting queued event")
	}

	if got := c.DroppedEvents(); got == 0 {
		t.Fatalf("expected dropped events > 0")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(dropped) == 0 {
		t.Fatalf("expected drop handler to be called")
	}
}

func TestErrorOverflowDropNewest(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(
		transport,
		WithErrorBuffer(1),
		WithErrorOverflowPolicy(OverflowPolicyDropNewest),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Close(context.Background())
	})

	transport.readCh <- []byte(`{"corrId":"100","resp":{"type":"cmdOk"}}`)
	transport.readCh <- []byte(`{"corrId":"101","resp":{"type":"cmdOk"}}`)

	deadline := time.After(2 * time.Second)
	for {
		if c.DroppedErrors() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting dropped error counter")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	select {
	case <-c.Errors():
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting queued error")
	}

	if got := c.DroppedErrors(); got == 0 {
		t.Fatalf("expected dropped errors > 0")
	}
}
