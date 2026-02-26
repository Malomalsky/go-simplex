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

func TestCreateContactInvitation(t *testing.T) {
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
				"type":"invitation",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"connLinkInvitation":{"connFullLink":"smp://full","connShortLink":"smp://short"},
				"connection":{}
			}
		}`)
		close(done)
	}()

	link, err := c.CreateContactInvitation(ctx, 1, false)
	if err != nil {
		t.Fatalf("CreateContactInvitation: %v", err)
	}
	<-done

	if link != "smp://short" {
		t.Fatalf("unexpected invitation link: %q", link)
	}
}

func TestAcceptContactRequest(t *testing.T) {
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
				"type":"acceptingContactRequest",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"contact":{"contactId":42,"profile":{"displayName":"alice"}}
			}
		}`)
		close(done)
	}()

	if err := c.AcceptContactRequest(ctx, 1001); err != nil {
		t.Fatalf("AcceptContactRequest: %v", err)
	}
	<-done
}

func TestRejectContactRequest(t *testing.T) {
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
				"type":"contactRequestRejected",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"contactRequest":{"id":1001}
			}
		}`)
		close(done)
	}()

	if err := c.RejectContactRequest(ctx, 1001); err != nil {
		t.Fatalf("RejectContactRequest: %v", err)
	}
	<-done
}

func TestAddGroupMember(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"sentGroupInvitation","user":{"userId":1},"groupInfo":{},"contact":{"contactId":42}}}`)
		close(done)
	}()

	if err := c.AddGroupMember(ctx, 7, 42, "member"); err != nil {
		t.Fatalf("AddGroupMember: %v", err)
	}
	<-done
}

func TestJoinGroup(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"userAcceptedGroupSent","user":{"userId":1},"groupInfo":{},"hostContact":{"contactId":42}}}`)
		close(done)
	}()

	if err := c.JoinGroup(ctx, 7); err != nil {
		t.Fatalf("JoinGroup: %v", err)
	}
	<-done
}

func TestAcceptGroupMember(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"memberAccepted","user":{"userId":1},"groupInfo":{},"member":{}}}`)
		close(done)
	}()

	if err := c.AcceptGroupMember(ctx, 7, 1001, "member"); err != nil {
		t.Fatalf("AcceptGroupMember: %v", err)
	}
	<-done
}

func TestSetGroupMembersRole(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"membersRoleUser","user":{"userId":1},"groupInfo":{},"members":[],"toRole":"member"}}`)
		close(done)
	}()

	if err := c.SetGroupMembersRole(ctx, 7, []int64{1001, 1002}, "member"); err != nil {
		t.Fatalf("SetGroupMembersRole: %v", err)
	}
	<-done
}

func TestBlockGroupMembersForAll(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"membersBlockedForAllUser","user":{"userId":1},"groupInfo":{},"members":[],"blocked":true}}`)
		close(done)
	}()

	if err := c.BlockGroupMembersForAll(ctx, 7, []int64{1001}, true); err != nil {
		t.Fatalf("BlockGroupMembersForAll: %v", err)
	}
	<-done
}

func TestRemoveGroupMembers(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"userDeletedMembers","user":{"userId":1},"groupInfo":{},"members":[],"withMessages":false}}`)
		close(done)
	}()

	if err := c.RemoveGroupMembers(ctx, 7, []int64{1001}, false); err != nil {
		t.Fatalf("RemoveGroupMembers: %v", err)
	}
	<-done
}

func TestLeaveGroup(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"leftMemberUser","user":{"userId":1},"groupInfo":{}}}`)
		close(done)
	}()

	if err := c.LeaveGroup(ctx, 7); err != nil {
		t.Fatalf("LeaveGroup: %v", err)
	}
	<-done
}

func TestListGroupMembers(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupMembers","user":{"userId":1},"group":{"members":[{"memberId":1}]}}}`)
		close(done)
	}()

	groupRaw, err := c.ListGroupMembers(ctx, 7)
	if err != nil {
		t.Fatalf("ListGroupMembers: %v", err)
	}
	<-done

	if string(groupRaw) == "" {
		t.Fatalf("unexpected empty group payload")
	}
}

func TestCreateGroup(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupCreated","user":{"userId":1},"groupInfo":{"groupId":7}}}`)
		close(done)
	}()

	groupRaw, err := c.CreateGroup(ctx, 1, false, map[string]any{"displayName": "support"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	<-done

	if string(groupRaw) == "" {
		t.Fatalf("unexpected empty group payload")
	}
}

func TestUpdateGroupProfile(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupUpdated","user":{"userId":1},"fromGroup":{},"toGroup":{"groupId":7},"member_":null}}`)
		close(done)
	}()

	groupRaw, err := c.UpdateGroupProfile(ctx, 7, map[string]any{"displayName": "new"})
	if err != nil {
		t.Fatalf("UpdateGroupProfile: %v", err)
	}
	<-done

	if string(groupRaw) == "" {
		t.Fatalf("unexpected empty group payload")
	}
}

func TestCreateGroupLink(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupLinkCreated","user":{"userId":1},"groupInfo":{},"groupLink":{"uri":"smp://group"}}}`)
		close(done)
	}()

	linkRaw, err := c.CreateGroupLink(ctx, 7, "member")
	if err != nil {
		t.Fatalf("CreateGroupLink: %v", err)
	}
	<-done

	if string(linkRaw) == "" {
		t.Fatalf("unexpected empty link payload")
	}
}

func TestSetGroupLinkMemberRole(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupLink","user":{"userId":1},"groupInfo":{},"groupLink":{"uri":"smp://group"}}}`)
		close(done)
	}()

	linkRaw, err := c.SetGroupLinkMemberRole(ctx, 7, "admin")
	if err != nil {
		t.Fatalf("SetGroupLinkMemberRole: %v", err)
	}
	<-done

	if string(linkRaw) == "" {
		t.Fatalf("unexpected empty link payload")
	}
}

func TestDeleteGroupLink(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupLinkDeleted","user":{"userId":1},"groupInfo":{}}}`)
		close(done)
	}()

	if err := c.DeleteGroupLink(ctx, 7); err != nil {
		t.Fatalf("DeleteGroupLink: %v", err)
	}
	<-done
}

func TestGetGroupLink(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupLink","user":{"userId":1},"groupInfo":{},"groupLink":{"uri":"smp://group"}}}`)
		close(done)
	}()

	linkRaw, err := c.GetGroupLink(ctx, 7)
	if err != nil {
		t.Fatalf("GetGroupLink: %v", err)
	}
	<-done

	if string(linkRaw) == "" {
		t.Fatalf("unexpected empty link payload")
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
