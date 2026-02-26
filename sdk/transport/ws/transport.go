package ws

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	HandshakeTimeout time.Duration
	ReadLimit        int64
	RequireWSS       bool
	TLSMinVersion    uint16
}

type Option func(*Config)

func WithHandshakeTimeout(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.HandshakeTimeout = d
		}
	}
}

func WithReadLimit(limit int64) Option {
	return func(c *Config) {
		if limit > 0 {
			c.ReadLimit = limit
		}
	}
}

func WithRequireWSS(require bool) Option {
	return func(c *Config) {
		c.RequireWSS = require
	}
}

func WithTLSMinVersion(version uint16) Option {
	return func(c *Config) {
		if version > 0 {
			c.TLSMinVersion = version
		}
	}
}

func defaultConfig() Config {
	return Config{
		HandshakeTimeout: 10 * time.Second,
		ReadLimit:        16 << 20,
		TLSMinVersion:    tls.VersionTLS12,
	}
}

type Transport struct {
	conn *websocket.Conn
	wmu  sync.Mutex

	closeOnce sync.Once
	closeErr  error
}

func Dial(ctx context.Context, endpoint string, opts ...Option) (*Transport, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("url is required")
	}
	parsedURL, err := urlpkg.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse websocket url: %w", err)
	}
	switch parsedURL.Scheme {
	case "ws", "wss":
	default:
		return nil, fmt.Errorf("unsupported websocket scheme: %s", parsedURL.Scheme)
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.RequireWSS && parsedURL.Scheme != "wss" {
		return nil, fmt.Errorf("wss is required, got scheme %q", parsedURL.Scheme)
	}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: cfg.HandshakeTimeout,
		TLSClientConfig: &tls.Config{
			MinVersion: cfg.TLSMinVersion,
		},
	}
	conn, resp, err := dialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("dial websocket: %w (status %s)", err, resp.Status)
		}
		return nil, fmt.Errorf("dial websocket: %w", err)
	}
	if cfg.ReadLimit > 0 {
		conn.SetReadLimit(cfg.ReadLimit)
	}
	return &Transport{conn: conn}, nil
}

func (t *Transport) Read(ctx context.Context) ([]byte, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if err := t.conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("set read deadline: %w", err)
		}
	} else {
		if err := t.conn.SetReadDeadline(time.Time{}); err != nil {
			return nil, fmt.Errorf("reset read deadline: %w", err)
		}
	}

	msgType, payload, err := t.conn.ReadMessage()
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}
	if msgType != websocket.TextMessage {
		return nil, fmt.Errorf("unexpected websocket frame type %d", msgType)
	}
	return payload, nil
}

func (t *Transport) Write(ctx context.Context, payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("payload is empty")
	}

	t.wmu.Lock()
	defer t.wmu.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		if err := t.conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("set write deadline: %w", err)
		}
	} else {
		if err := t.conn.SetWriteDeadline(time.Time{}); err != nil {
			return fmt.Errorf("reset write deadline: %w", err)
		}
	}

	if err := t.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return fmt.Errorf("write websocket frame: %w", err)
		}
	}
	return nil
}

func (t *Transport) Close() error {
	t.closeOnce.Do(func() {
		t.closeErr = t.conn.Close()
	})
	return t.closeErr
}
