package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/transport/ws"
)

type Runner interface {
	Run(ctx context.Context) error
}

type DialFunc func(ctx context.Context) (*client.Client, error)
type RunnerFactory func(cli *client.Client) (Runner, error)

type ReconnectOption func(*ReconnectConfig)
type ReconnectErrorHandler func(error)

type ReconnectConfig struct {
	MinBackoff             time.Duration
	MaxBackoff             time.Duration
	MaxConsecutiveFailures int
	StablePeriod           time.Duration
	CloseTimeout           time.Duration
	OnError                ReconnectErrorHandler
}

func defaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MinBackoff:   1 * time.Second,
		MaxBackoff:   30 * time.Second,
		StablePeriod: 30 * time.Second,
		CloseTimeout: 5 * time.Second,
		OnError:      func(error) {},
	}
}

func WithReconnectBackoff(minBackoff, maxBackoff time.Duration) ReconnectOption {
	return func(c *ReconnectConfig) {
		if minBackoff > 0 {
			c.MinBackoff = minBackoff
		}
		if maxBackoff > 0 {
			c.MaxBackoff = maxBackoff
		}
	}
}

func WithReconnectMaxConsecutiveFailures(max int) ReconnectOption {
	return func(c *ReconnectConfig) {
		if max >= 0 {
			c.MaxConsecutiveFailures = max
		}
	}
}

func WithReconnectStablePeriod(d time.Duration) ReconnectOption {
	return func(c *ReconnectConfig) {
		if d >= 0 {
			c.StablePeriod = d
		}
	}
}

func WithReconnectCloseTimeout(d time.Duration) ReconnectOption {
	return func(c *ReconnectConfig) {
		if d >= 0 {
			c.CloseTimeout = d
		}
	}
}

func WithReconnectErrorHandler(h ReconnectErrorHandler) ReconnectOption {
	return func(c *ReconnectConfig) {
		if h != nil {
			c.OnError = h
		}
	}
}

func RunWebSocketWithReconnect(
	ctx context.Context,
	url string,
	wsOptions []ws.Option,
	clientOptions []client.Option,
	build RunnerFactory,
	opts ...ReconnectOption,
) error {
	dial := func(dialCtx context.Context) (*client.Client, error) {
		return client.NewWebSocketWithOptions(dialCtx, url, wsOptions, clientOptions...)
	}
	return RunWithReconnect(ctx, dial, build, opts...)
}

func RunWithReconnect(ctx context.Context, dial DialFunc, build RunnerFactory, opts ...ReconnectOption) error {
	if ctx == nil {
		return fmt.Errorf("ctx is nil")
	}
	if dial == nil {
		return fmt.Errorf("dial is nil")
	}
	if build == nil {
		return fmt.Errorf("build is nil")
	}

	cfg := defaultReconnectConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.MaxBackoff < cfg.MinBackoff {
		cfg.MaxBackoff = cfg.MinBackoff
	}

	failures := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		cli, err := dial(ctx)
		if err != nil {
			failures++
			cfg.OnError(fmt.Errorf("dial client: %w", err))
			if err := waitReconnectBackoff(ctx, cfg, failures); err != nil {
				return err
			}
			continue
		}

		runner, err := build(cli)
		if err != nil {
			closeClient(cli, cfg.CloseTimeout)
			return fmt.Errorf("build runner: %w", err)
		}

		started := time.Now()
		runErr := runner.Run(ctx)
		closeClient(cli, cfg.CloseTimeout)

		if err := ctx.Err(); err != nil {
			return err
		}

		if runErr == nil {
			runErr = errors.New("runner stopped without error")
		}
		cfg.OnError(fmt.Errorf("runner stopped: %w", runErr))

		if cfg.StablePeriod > 0 && time.Since(started) >= cfg.StablePeriod {
			failures = 1
		} else {
			failures++
		}
		if err := waitReconnectBackoff(ctx, cfg, failures); err != nil {
			return err
		}
	}
}

func waitReconnectBackoff(ctx context.Context, cfg ReconnectConfig, failures int) error {
	if cfg.MaxConsecutiveFailures > 0 && failures >= cfg.MaxConsecutiveFailures {
		return fmt.Errorf("max consecutive failures reached: %d", failures)
	}
	delay := reconnectDelay(cfg.MinBackoff, cfg.MaxBackoff, failures)
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func reconnectDelay(minBackoff, maxBackoff time.Duration, failures int) time.Duration {
	if minBackoff <= 0 {
		return 0
	}
	if maxBackoff > 0 && maxBackoff < minBackoff {
		maxBackoff = minBackoff
	}
	if failures <= 0 {
		return 0
	}
	delay := minBackoff
	for i := 1; i < failures; i++ {
		if maxBackoff > 0 && delay >= maxBackoff/2 {
			delay = maxBackoff
			return delay
		}
		delay *= 2
	}
	if maxBackoff > 0 && delay > maxBackoff {
		return maxBackoff
	}
	return delay
}

func closeClient(cli *client.Client, timeout time.Duration) {
	if cli == nil {
		return
	}
	if timeout <= 0 {
		_ = cli.Close(context.Background())
		return
	}
	closeCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = cli.Close(closeCtx)
}
