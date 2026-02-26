package client

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestGetActiveUser(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
			Cmd    string `json:"cmd"`
		}
		_ = json.Unmarshal(rawReq, &req)
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"activeUser","user":{"userId":11,"profile":{"displayName":"bot"}}}}`)
		close(done)
	}()

	user, err := c.GetActiveUser(ctx)
	if err != nil {
		t.Fatalf("GetActiveUser: %v", err)
	}
	<-done

	if user.UserID != 11 {
		t.Fatalf("unexpected user id: %d", user.UserID)
	}
	if user.Profile.DisplayName != "bot" {
		t.Fatalf("unexpected display name: %q", user.Profile.DisplayName)
	}
}

func TestEnsureUserAddress(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		// first command: show address -> not found
		rawReq1 := <-transport.writeCh
		var req1 struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(rawReq1, &req1)
		transport.readCh <- []byte(`{"corrId":"` + req1.CorrID + `","resp":{"type":"chatCmdError","chatError":{"type":"errorStore","storeError":{"type":"userContactLinkNotFound"}}}}`)

		// second command: create address
		rawReq2 := <-transport.writeCh
		var req2 struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(rawReq2, &req2)
		transport.readCh <- []byte(`{"corrId":"` + req2.CorrID + `","resp":{"type":"userContactLinkCreated","user":{"userId":1},"connLinkContact":{"connFullLink":"smp://full","connShortLink":"smp://short"}}}`)
		close(done)
	}()

	addr, err := c.EnsureUserAddress(ctx, 1)
	if err != nil {
		t.Fatalf("EnsureUserAddress: %v", err)
	}
	<-done

	if addr != "smp://short" {
		t.Fatalf("unexpected address: %q", addr)
	}
}

func TestDeleteUserAddress(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(rawReq, &req)
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"userContactLinkDeleted","user":{"userId":1}}}`)
		close(done)
	}()

	if err := c.DeleteUserAddress(ctx, 1); err != nil {
		t.Fatalf("DeleteUserAddress: %v", err)
	}
	<-done
}

func TestEnsureUserAddressPropagatesUnexpectedError(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(rawReq, &req)
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"chatCmdError","chatError":{"type":"error"}}}`)
	}()

	_, err = c.EnsureUserAddress(ctx, 1)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestListContacts(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(rawReq, &req)
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"contactsList",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"contacts":[
					{"contactId":42,"profile":{"displayName":"alice"}},
					{"contactId":43,"profile":{"displayName":"bob"}}
				]
			}
		}`)
		close(done)
	}()

	contacts, err := c.ListContacts(ctx, 1)
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	<-done

	if len(contacts) != 2 {
		t.Fatalf("unexpected contacts count: %d", len(contacts))
	}
	if contacts[0].ContactID != 42 || contacts[0].Profile.DisplayName != "alice" {
		t.Fatalf("unexpected first contact: %+v", contacts[0])
	}
}

func TestListGroups(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(rawReq, &req)
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"groupsList",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"groups":[{"groupId":7},{"groupId":8}]
			}
		}`)
		close(done)
	}()

	contactID := int64(42)
	groups, err := c.ListGroups(ctx, 1, &contactID, "support")
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	<-done

	if len(groups) != 2 {
		t.Fatalf("unexpected groups count: %d", len(groups))
	}
	if string(groups[0]) == "" {
		t.Fatalf("unexpected empty group payload")
	}
}

func TestSendTextMessage(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
			Cmd    string `json:"cmd"`
		}
		_ = json.Unmarshal(rawReq, &req)
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"newChatItems","chatItems":[]}}`)
		close(done)
	}()

	if err := c.SendTextMessage(ctx, "@42", "hello"); err != nil {
		t.Fatalf("SendTextMessage: %v", err)
	}
	<-done
}

func TestSendTextToContact(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan string, 1)
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
			Cmd    string `json:"cmd"`
		}
		_ = json.Unmarshal(rawReq, &req)
		done <- req.Cmd
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"newChatItems","chatItems":[]}}`)
	}()

	if err := c.SendTextToContact(ctx, 42, "hello"); err != nil {
		t.Fatalf("SendTextToContact: %v", err)
	}

	cmd := <-done
	if cmd == "" || cmd[:10] != "/_send @42" {
		t.Fatalf("unexpected contact cmd: %q", cmd)
	}
}

func TestSendTextToGroup(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan string, 1)
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
			Cmd    string `json:"cmd"`
		}
		_ = json.Unmarshal(rawReq, &req)
		done <- req.Cmd
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"newChatItems","chatItems":[]}}`)
	}()

	if err := c.SendTextToGroup(ctx, 7, "hello"); err != nil {
		t.Fatalf("SendTextToGroup: %v", err)
	}

	cmd := <-done
	if cmd == "" || cmd[:9] != "/_send #7" {
		t.Fatalf("unexpected group cmd: %q", cmd)
	}
}
