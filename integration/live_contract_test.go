//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/command"
)

const (
	envSimplexWSURL = "SIMPLEX_WS_URL"

	envSimplexTestContactID         = "SIMPLEX_TEST_CONTACT_ID"
	envSimplexTestContactChatItemID = "SIMPLEX_TEST_CONTACT_CHAT_ITEM_ID"
	envSimplexTestGroupID           = "SIMPLEX_TEST_GROUP_ID"
	envSimplexTestGroupChatItemID   = "SIMPLEX_TEST_GROUP_CHAT_ITEM_ID"
	envSimplexTestFileID            = "SIMPLEX_TEST_FILE_ID"
)

func liveWSURL(t *testing.T) string {
	t.Helper()
	url := strings.TrimSpace(os.Getenv(envSimplexWSURL))
	if url == "" {
		t.Skipf("set %s to run live contract tests", envSimplexWSURL)
	}
	return url
}

func mustEnvInt64(t *testing.T, envName string) int64 {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		t.Skipf("set %s to run this live test", envName)
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		t.Skipf("%s must be int64, got %q: %v", envName, raw, err)
	}
	if value <= 0 {
		t.Skipf("%s must be > 0, got %d", envName, value)
	}
	return value
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
	return context.WithTimeout(context.Background(), 20*time.Second)
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

func TestLiveListContactsAndGroups(t *testing.T) {
	wsURL := liveWSURL(t)
	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	user, err := cli.GetActiveUser(ctx)
	if err != nil {
		t.Fatalf("GetActiveUser: %v", err)
	}

	contacts, err := cli.ListContacts(ctx, user.UserID)
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	if contacts == nil {
		t.Fatalf("ListContacts: got nil slice")
	}

	groups, err := cli.ListGroupsTyped(ctx, user.UserID, nil, "")
	if err != nil {
		t.Fatalf("ListGroupsTyped: %v", err)
	}
	if groups == nil {
		t.Fatalf("ListGroupsTyped: got nil slice")
	}
}

func TestLiveContactMessageFlow(t *testing.T) {
	wsURL := liveWSURL(t)
	contactID := mustEnvInt64(t, envSimplexTestContactID)
	chatItemID := mustEnvInt64(t, envSimplexTestContactChatItemID)

	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	text := fmt.Sprintf("contract-contact-%d", time.Now().UnixNano())
	if err := cli.SendTextToContact(ctx, contactID, text); err != nil {
		t.Fatalf("SendTextToContact: %v", err)
	}

	updated := text + "-updated"
	summary, err := cli.UpdateTextMessageInContact(ctx, contactID, chatItemID, updated, false)
	if err != nil {
		t.Fatalf("UpdateTextMessageInContact: %v", err)
	}
	if summary == nil {
		t.Fatalf("UpdateTextMessageInContact: nil summary")
	}

	reaction := map[string]any{"type": "emoji", "emoji": ":thumbsup:"}
	if _, err := cli.AddChatItemReaction(ctx, command.DirectRef(contactID), chatItemID, reaction); err != nil {
		t.Fatalf("AddChatItemReaction: %v", err)
	}
	if _, err := cli.RemoveChatItemReaction(ctx, command.DirectRef(contactID), chatItemID, reaction); err != nil {
		t.Fatalf("RemoveChatItemReaction: %v", err)
	}

	deleted, err := cli.DeleteChatItemsInContact(ctx, contactID, []int64{chatItemID}, client.CIDeleteModeInternal)
	if err != nil {
		t.Fatalf("DeleteChatItemsInContact: %v", err)
	}
	if deleted == nil {
		t.Fatalf("DeleteChatItemsInContact: nil response")
	}
}

func TestLiveGroupMessageFlow(t *testing.T) {
	wsURL := liveWSURL(t)
	groupID := mustEnvInt64(t, envSimplexTestGroupID)
	chatItemID := mustEnvInt64(t, envSimplexTestGroupChatItemID)

	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	text := fmt.Sprintf("contract-group-%d", time.Now().UnixNano())
	if err := cli.SendTextToGroup(ctx, groupID, text); err != nil {
		t.Fatalf("SendTextToGroup: %v", err)
	}

	updated := text + "-updated"
	summary, err := cli.UpdateTextMessageInGroup(ctx, groupID, chatItemID, updated, false)
	if err != nil {
		t.Fatalf("UpdateTextMessageInGroup: %v", err)
	}
	if summary == nil {
		t.Fatalf("UpdateTextMessageInGroup: nil summary")
	}

	reaction := map[string]any{"type": "emoji", "emoji": ":eyes:"}
	if _, err := cli.AddChatItemReaction(ctx, command.GroupRef(groupID), chatItemID, reaction); err != nil {
		t.Fatalf("AddChatItemReaction(group): %v", err)
	}
	if _, err := cli.RemoveChatItemReaction(ctx, command.GroupRef(groupID), chatItemID, reaction); err != nil {
		t.Fatalf("RemoveChatItemReaction(group): %v", err)
	}

	deleted, err := cli.DeleteChatItemsInGroup(ctx, groupID, []int64{chatItemID}, client.CIDeleteModeInternal)
	if err != nil {
		t.Fatalf("DeleteChatItemsInGroup: %v", err)
	}
	if deleted == nil {
		t.Fatalf("DeleteChatItemsInGroup: nil response")
	}
}

func TestLiveFileCommands(t *testing.T) {
	wsURL := liveWSURL(t)
	fileID := mustEnvInt64(t, envSimplexTestFileID)
	cli := newLiveClient(t, wsURL, client.WithStrictResponses(false))

	ctx, cancel := opCtx(t)
	defer cancel()

	rcv, err := cli.ReceiveFile(ctx, fileID, client.ReceiveFileOptions{})
	if err != nil {
		t.Fatalf("ReceiveFile: %v", err)
	}
	if rcv == nil || rcv.ResponseType == "" {
		t.Fatalf("ReceiveFile: invalid response summary: %+v", rcv)
	}

	cancelSummary, err := cli.CancelFile(ctx, fileID)
	if err != nil {
		t.Fatalf("CancelFile: %v", err)
	}
	if cancelSummary == nil || cancelSummary.ResponseType == "" {
		t.Fatalf("CancelFile: invalid response summary: %+v", cancelSummary)
	}
}
