package bot

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
)

type fakeRunner struct {
	run func(ctx context.Context) error
}

func (f fakeRunner) Run(ctx context.Context) error {
	return f.run(ctx)
}

func TestRunWithReconnectNilArguments(t *testing.T) {
	t.Parallel()

	err := RunWithReconnect(context.Background(), nil, nil)
	if err == nil {
		t.Fatalf("expected argument validation error")
	}
}

func TestRunWithReconnectBuildErrorClosesClient(t *testing.T) {
	t.Parallel()

	tr := newMockTransport()
	dial := func(ctx context.Context) (*client.Client, error) {
		return client.New(tr)
	}
	build := func(cli *client.Client) (Runner, error) {
		return nil, errors.New("broken builder")
	}

	err := RunWithReconnect(context.Background(), dial, build)
	if err == nil || !strings.Contains(err.Error(), "build runner") {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-tr.closed:
	case <-time.After(2 * time.Second):
		t.Fatalf("client transport should be closed on build error")
	}
}

func TestRunWithReconnectMaxConsecutiveFailures(t *testing.T) {
	t.Parallel()

	var dials int32
	dial := func(ctx context.Context) (*client.Client, error) {
		atomic.AddInt32(&dials, 1)
		tr := newMockTransport()
		return client.New(tr)
	}
	build := func(cli *client.Client) (Runner, error) {
		return fakeRunner{
			run: func(ctx context.Context) error {
				return errors.New("runtime boom")
			},
		}, nil
	}

	err := RunWithReconnect(
		context.Background(),
		dial,
		build,
		WithReconnectBackoff(1*time.Millisecond, 2*time.Millisecond),
		WithReconnectMaxConsecutiveFailures(2),
	)
	if err == nil {
		t.Fatalf("expected max consecutive failures error")
	}
	if !strings.Contains(err.Error(), "max consecutive failures reached") {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&dials); got < 2 {
		t.Fatalf("expected multiple reconnect attempts, got %d", got)
	}
}

func TestReconnectDelay(t *testing.T) {
	t.Parallel()

	if d := reconnectDelay(0, 0, 1); d != 0 {
		t.Fatalf("expected zero delay for disabled min backoff, got %s", d)
	}
	if d := reconnectDelay(10*time.Millisecond, 50*time.Millisecond, 1); d != 10*time.Millisecond {
		t.Fatalf("unexpected delay for first failure: %s", d)
	}
	if d := reconnectDelay(10*time.Millisecond, 50*time.Millisecond, 2); d != 20*time.Millisecond {
		t.Fatalf("unexpected delay for second failure: %s", d)
	}
	if d := reconnectDelay(10*time.Millisecond, 50*time.Millisecond, 4); d != 50*time.Millisecond {
		t.Fatalf("expected capped delay, got %s", d)
	}
}
