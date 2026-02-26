package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Malomalsky/go-simplex/sdk/command"
	"github.com/Malomalsky/go-simplex/sdk/protocol"
)

var ErrClosed = errors.New("client closed")

type UnexpectedResponseTypeError struct {
	Command      string
	ResponseType string
	Expected     []string
}

func (e *UnexpectedResponseTypeError) Error() string {
	if e == nil {
		return "unexpected response type"
	}
	return fmt.Sprintf(
		"unexpected response type for %s: got %q, expected one of [%s]",
		e.Command,
		e.ResponseType,
		strings.Join(e.Expected, ", "),
	)
}

type Transport interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, payload []byte) error
	Close() error
}

type Config struct {
	EventBuffer int
	ErrorBuffer int
}

type Option func(*Config)

func WithEventBuffer(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.EventBuffer = size
		}
	}
}

func WithErrorBuffer(size int) Option {
	return func(c *Config) {
		if size > 0 {
			c.ErrorBuffer = size
		}
	}
}

func defaultConfig() Config {
	return Config{
		EventBuffer: 128,
		ErrorBuffer: 16,
	}
}

type pendingResult struct {
	msg protocol.Message
	err error
}

type Client struct {
	transport Transport

	events chan protocol.Message
	errs   chan error

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	pendingMu sync.Mutex
	pending   map[string]chan pendingResult

	closeOnce sync.Once
	nextID    uint64
}

func New(transport Transport, opts ...Option) (*Client, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is required")
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		transport: transport,
		events:    make(chan protocol.Message, cfg.EventBuffer),
		errs:      make(chan error, cfg.ErrorBuffer),
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
		pending:   make(map[string]chan pendingResult),
	}
	go c.readLoop()
	return c, nil
}

func (c *Client) Events() <-chan protocol.Message {
	return c.events
}

func (c *Client) Errors() <-chan error {
	return c.errs
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) SendRaw(ctx context.Context, cmd string) (protocol.Message, error) {
	if cmd == "" {
		return protocol.Message{}, fmt.Errorf("cmd is required")
	}
	if err := c.ensureOpen(); err != nil {
		return protocol.Message{}, err
	}

	corrID := strconv.FormatUint(atomic.AddUint64(&c.nextID, 1), 10)
	req := protocol.CommandRequest{CorrID: corrID, Cmd: cmd}

	payload, err := protocol.EncodeRequest(req)
	if err != nil {
		return protocol.Message{}, err
	}

	wait := make(chan pendingResult, 1)
	c.setPending(corrID, wait)

	if err := c.transport.Write(ctx, payload); err != nil {
		c.removePending(corrID)
		return protocol.Message{}, fmt.Errorf("write request: %w", err)
	}

	select {
	case res := <-wait:
		if res.err != nil {
			return protocol.Message{}, res.err
		}
		return res.msg, nil
	case <-ctx.Done():
		c.removePending(corrID)
		return protocol.Message{}, ctx.Err()
	case <-c.done:
		c.removePending(corrID)
		return protocol.Message{}, ErrClosed
	}
}

func (c *Client) Send(ctx context.Context, req command.Request) (protocol.Message, error) {
	if req == nil {
		return protocol.Message{}, fmt.Errorf("request is nil")
	}

	cmdString, err := safeCommandString(req)
	if err != nil {
		return protocol.Message{}, err
	}

	msg, err := c.SendRaw(ctx, cmdString)
	if err != nil {
		return protocol.Message{}, err
	}

	expected := command.ExpectedResponseTypes(req)
	if len(expected) == 0 {
		return msg, nil
	}
	if containsString(expected, msg.Resp.Type) {
		return msg, nil
	}

	return protocol.Message{}, &UnexpectedResponseTypeError{
		Command:      fmt.Sprintf("%T", req),
		ResponseType: msg.Resp.Type,
		Expected:     append([]string(nil), expected...),
	}
}

func (c *Client) Close(ctx context.Context) error {
	c.close()
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		c.cancel()
		_ = c.transport.Close()
		c.failPending(ErrClosed)
	})
}

func (c *Client) readLoop() {
	defer close(c.done)
	defer close(c.events)
	defer close(c.errs)

	for {
		payload, err := c.transport.Read(c.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			c.emitErr(fmt.Errorf("read transport: %w", err))
			c.close()
			return
		}

		msg, err := protocol.DecodeMessage(payload)
		if err != nil {
			c.emitErr(err)
			continue
		}

		if msg.IsEvent() {
			c.emitEvent(msg)
			continue
		}

		if !c.resolvePending(msg) {
			c.emitErr(fmt.Errorf("no pending command for corrId=%s", msg.CorrID))
		}
	}
}

func (c *Client) resolvePending(msg protocol.Message) bool {
	c.pendingMu.Lock()
	wait, ok := c.pending[msg.CorrID]
	if ok {
		delete(c.pending, msg.CorrID)
	}
	c.pendingMu.Unlock()
	if !ok {
		return false
	}
	wait <- pendingResult{msg: msg}
	return true
}

func (c *Client) setPending(corrID string, wait chan pendingResult) {
	c.pendingMu.Lock()
	c.pending[corrID] = wait
	c.pendingMu.Unlock()
}

func (c *Client) removePending(corrID string) {
	c.pendingMu.Lock()
	delete(c.pending, corrID)
	c.pendingMu.Unlock()
}

func (c *Client) failPending(err error) {
	c.pendingMu.Lock()
	pending := c.pending
	c.pending = make(map[string]chan pendingResult)
	c.pendingMu.Unlock()

	for _, wait := range pending {
		wait <- pendingResult{err: err}
	}
}

func (c *Client) emitEvent(msg protocol.Message) {
	select {
	case c.events <- msg:
	case <-c.done:
	case <-c.ctx.Done():
	}
}

func (c *Client) emitErr(err error) {
	if err == nil {
		return
	}
	select {
	case c.errs <- err:
	case <-c.done:
	case <-c.ctx.Done():
	}
}

func (c *Client) ensureOpen() error {
	select {
	case <-c.done:
		return ErrClosed
	default:
		return nil
	}
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func safeCommandString(req command.Request) (cmd string, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("build command string panic: %v", rec)
		}
	}()
	cmd = req.CommandString()
	return cmd, err
}
