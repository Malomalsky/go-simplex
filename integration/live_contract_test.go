//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/command"
)

const envSimplexWSURL = "SIMPLEX_WS_URL"

func liveWSURL(t *testing.T) string {
	t.Helper()
	url := strings.TrimSpace(os.Getenv(envSimplexWSURL))
	if url == "" {
		t.Skipf("set %s to run live contract tests", envSimplexWSURL)
	}
	return url
}

func newLiveClient(t *testing.T, wsURL string, opts ...client.Option) *client.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cli, err := client.NewWebSocket(ctx, wsURL, opts...)
	if err != nil {
		t.Fatalf("connect websocket: %v", err)
	}
	t.Cleanup(func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := cli.Close(closeCtx); err != nil {
			t.Fatalf("close client: %v", err)
		}
	})
	return cli
}

func opCtx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 15*time.Second)
}

func TestLiveGetActiveUser(t *testing.T) {
	wsURL := liveWSURL(t)
	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	user, err := cli.GetActiveUser(ctx)
	if err != nil {
		t.Fatalf("GetActiveUser: %v", err)
	}
	if user == nil {
		t.Fatalf("GetActiveUser: got nil user")
	}
	if user.UserID <= 0 {
		t.Fatalf("GetActiveUser: invalid user id %d", user.UserID)
	}
}

func TestLiveBootstrapBot(t *testing.T) {
	wsURL := liveWSURL(t)
	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	boot, err := cli.BootstrapBot(ctx)
	if err != nil {
		t.Fatalf("BootstrapBot: %v", err)
	}
	if boot == nil {
		t.Fatalf("BootstrapBot: got nil result")
	}
	if boot.User == nil || boot.User.UserID <= 0 {
		t.Fatalf("BootstrapBot: invalid user: %+v", boot.User)
	}
	if strings.TrimSpace(boot.Address) == "" {
		t.Fatalf("BootstrapBot: empty address")
	}
}

func TestLiveTypedSenderParity(t *testing.T) {
	wsURL := liveWSURL(t)
	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	typed, err := cli.SendShowActiveUser(ctx, command.ShowActiveUser{})
	if err != nil {
		t.Fatalf("SendShowActiveUser: %v", err)
	}
	if typed.ActiveUser == nil {
		t.Fatalf("SendShowActiveUser: missing activeUser payload (resp.type=%s)", typed.Message.Resp.Type)
	}

	helper, err := cli.GetActiveUser(ctx)
	if err != nil {
		t.Fatalf("GetActiveUser: %v", err)
	}

	if helper.UserID != typed.ActiveUser.User.UserID {
		t.Fatalf("active user mismatch: helper=%d typed=%d", helper.UserID, typed.ActiveUser.User.UserID)
	}
}

func TestLiveListUsersTyped(t *testing.T) {
	wsURL := liveWSURL(t)
	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	users, err := cli.ListUsersTyped(ctx)
	if err != nil {
		t.Fatalf("ListUsersTyped: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("ListUsersTyped: expected at least one user")
	}
}
