package client

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Malomalsky/go-simplex/sdk/types"
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

func TestListGroupsTyped(t *testing.T) {
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
				"groups":[{"groupInfo":{"groupId":7}},{"groupInfo":{"groupId":8}}]
			}
		}`)
		close(done)
	}()

	groups, err := c.ListGroupsTyped(ctx, 1, nil, "")
	if err != nil {
		t.Fatalf("ListGroupsTyped: %v", err)
	}
	<-done

	if len(groups) != 2 {
		t.Fatalf("unexpected groups count: %d", len(groups))
	}
	if groups[0].GroupInfo.GroupId != 7 {
		t.Fatalf("unexpected group id: %d", groups[0].GroupInfo.GroupId)
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"groupMembers","user":{"userId":1},"group":{"members":[{"memberId":"m1"}]}}}`)
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

func TestListGroupMembersTyped(t *testing.T) {
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
				"type":"groupMembers",
				"user":{"userId":1},
				"group":{"groupInfo":{"groupId":7},"members":[{"memberId":"m1"}]}
			}
		}`)
		close(done)
	}()

	group, err := c.ListGroupMembersTyped(ctx, 7)
	if err != nil {
		t.Fatalf("ListGroupMembersTyped: %v", err)
	}
	<-done

	if group.GroupInfo.GroupId != 7 {
		t.Fatalf("unexpected group id: %d", group.GroupInfo.GroupId)
	}
	if len(group.Members) != 1 {
		t.Fatalf("unexpected members count: %d", len(group.Members))
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

func TestCreateUser(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"activeUser","user":{"userId":11,"profile":{"displayName":"bot2"}}}}`)
		close(done)
	}()

	user, err := c.CreateUser(ctx, map[string]any{"displayName": "bot2"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	<-done

	if user.UserID != 11 || user.Profile.DisplayName != "bot2" {
		t.Fatalf("unexpected created user: %+v", user)
	}
}

func TestListUsers(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"usersList","users":[{"userId":1},{"userId":2}]}}`)
		close(done)
	}()

	users, err := c.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	<-done

	if len(users) != 2 {
		t.Fatalf("unexpected users count: %d", len(users))
	}
}

func TestListUsersTyped(t *testing.T) {
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
				"type":"usersList",
				"users":[{"user":{"userId":1},"unreadCount":2},{"user":{"userId":2},"unreadCount":3}]
			}
		}`)
		close(done)
	}()

	users, err := c.ListUsersTyped(ctx)
	if err != nil {
		t.Fatalf("ListUsersTyped: %v", err)
	}
	<-done

	if len(users) != 2 {
		t.Fatalf("unexpected users count: %d", len(users))
	}
	if users[0].UnreadCount != 2 {
		t.Fatalf("unexpected unread count: %d", users[0].UnreadCount)
	}
}

func TestSetActiveUser(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"activeUser","user":{"userId":2,"profile":{"displayName":"active"}}}}`)
		close(done)
	}()

	user, err := c.SetActiveUser(ctx, 2, nil)
	if err != nil {
		t.Fatalf("SetActiveUser: %v", err)
	}
	<-done

	if user.UserID != 2 {
		t.Fatalf("unexpected active user id: %d", user.UserID)
	}
}

func TestDeleteUser(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"cmdOk"}}`)
		close(done)
	}()

	if err := c.DeleteUser(ctx, 2, false, nil); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	<-done
}

func TestUpdateProfile(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"userProfileUpdated","user":{"userId":1},"fromProfile":{},"toProfile":{},"updateSummary":{}}}`)
		close(done)
	}()

	changed, err := c.UpdateProfile(ctx, 1, map[string]any{"displayName": "new"})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	<-done

	if !changed {
		t.Fatalf("expected changed=true")
	}
}

func TestSetContactPreferences(t *testing.T) {
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
		transport.readCh <- []byte(`{"corrId":"` + req.CorrID + `","resp":{"type":"contactPrefsUpdated","user":{"userId":1},"fromContact":{"contactId":7},"toContact":{"contactId":7}}}`)
		close(done)
	}()

	if err := c.SetContactPreferences(ctx, 7, map[string]any{"timedMessages": false}); err != nil {
		t.Fatalf("SetContactPreferences: %v", err)
	}
	<-done
}

func TestConnectPlan(t *testing.T) {
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
				"type":"connectionPlan",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"connLink":{"connFullLink":"smp://full","connShortLink":"smp://short"},
				"connectionPlan":{"plan":"contact"}
			}
		}`)
		close(done)
	}()

	plan, err := c.ConnectPlan(ctx, 1, "smp://invite")
	if err != nil {
		t.Fatalf("ConnectPlan: %v", err)
	}
	<-done

	if plan.ConnLink.PreferredLink() != "smp://short" {
		t.Fatalf("unexpected plan link: %+v", plan.ConnLink)
	}
}

func TestConnectWithPreparedLink(t *testing.T) {
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
				"type":"contactAlreadyExists",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"contact":{"contactId":42,"profile":{"displayName":"alice"}}
			}
		}`)
		close(done)
	}()

	link := "smp://invite"
	res, err := c.ConnectWithPreparedLink(ctx, 1, false, &link)
	if err != nil {
		t.Fatalf("ConnectWithPreparedLink: %v", err)
	}
	<-done

	if res.ResponseType != types.ResponseTypeContactAlreadyExists {
		t.Fatalf("unexpected response type: %s", res.ResponseType)
	}
	if res.ExistingContact == nil || res.ExistingContact.ContactID != 42 {
		t.Fatalf("unexpected existing contact: %+v", res.ExistingContact)
	}
}

func TestConnectWithLink(t *testing.T) {
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
				"type":"sentInvitation",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"connection":{"id":"conn-1"},
				"customUserProfile":{}
			}
		}`)
		close(done)
	}()

	link := "smp://invite"
	res, err := c.ConnectWithLink(ctx, &link)
	if err != nil {
		t.Fatalf("ConnectWithLink: %v", err)
	}
	<-done

	if res.ResponseType != types.ResponseTypeSentInvitation {
		t.Fatalf("unexpected response type: %s", res.ResponseType)
	}
	if string(res.Connection) == "" {
		t.Fatalf("expected non-empty connection payload")
	}
	if res.ConnectionInfo == nil {
		t.Fatalf("expected non-nil typed connection details")
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

func TestSendTextMessageWithOptions(t *testing.T) {
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

	ttl := int64(120)
	if err := c.SendTextMessageWithOptions(ctx, "@42", "hello", SendTextOptions{
		Live: true,
		TTL:  &ttl,
	}); err != nil {
		t.Fatalf("SendTextMessageWithOptions: %v", err)
	}

	cmd := <-done
	if !strings.Contains(cmd, " live=on") {
		t.Fatalf("expected live=on in command: %q", cmd)
	}
	if !strings.Contains(cmd, " ttl=120") {
		t.Fatalf("expected ttl=120 in command: %q", cmd)
	}
}

func TestSendTextMessageInvalidSendRef(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	err = c.SendTextMessage(context.Background(), "invalid", "hello")
	if err == nil {
		t.Fatalf("expected invalid sendRef error")
	}
	if !strings.Contains(err.Error(), "invalid sendRef") {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-transport.writeCh:
		t.Fatalf("invalid sendRef should not write command")
	default:
	}
}

func TestSendTextToContactWithOptions(t *testing.T) {
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

	if err := c.SendTextToContactWithOptions(ctx, 42, "hello", SendTextOptions{Live: true}); err != nil {
		t.Fatalf("SendTextToContactWithOptions: %v", err)
	}

	cmd := <-done
	if !strings.HasPrefix(cmd, "/_send @42") {
		t.Fatalf("unexpected contact cmd: %q", cmd)
	}
	if !strings.Contains(cmd, " live=on") {
		t.Fatalf("expected live=on in contact command: %q", cmd)
	}
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

func TestSendTextToGroupWithOptions(t *testing.T) {
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

	if err := c.SendTextToGroupWithOptions(ctx, 7, "hello", SendTextOptions{Live: true}); err != nil {
		t.Fatalf("SendTextToGroupWithOptions: %v", err)
	}

	cmd := <-done
	if !strings.HasPrefix(cmd, "/_send #7") {
		t.Fatalf("unexpected group cmd: %q", cmd)
	}
	if !strings.Contains(cmd, " live=on") {
		t.Fatalf("expected live=on in group command: %q", cmd)
	}
}

func TestUpdateTextMessage(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"chatItemUpdated",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItem":{
					"chatInfo":{"type":"direct","contact":{"contactId":42}},
					"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"updated"}}}
				}
			}
		}`)
	}()

	res, err := c.UpdateTextMessage(ctx, "@42", 7, "updated", true)
	if err != nil {
		t.Fatalf("UpdateTextMessage: %v", err)
	}
	cmd := <-done

	if res.ResponseType != types.ResponseTypeChatItemUpdated {
		t.Fatalf("unexpected response type: %s", res.ResponseType)
	}
	if !res.Updated {
		t.Fatalf("expected updated=true")
	}
	if !strings.HasPrefix(cmd, "/_update item @42 7") {
		t.Fatalf("unexpected update command: %q", cmd)
	}
	if !strings.Contains(cmd, " live=on") {
		t.Fatalf("expected live=on in update command: %q", cmd)
	}
}

func TestUpdateChatItemInvalidChatRef(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	_, err = c.UpdateChatItem(context.Background(), "bad", 7, map[string]any{
		"msgContent": map[string]any{"type": "text", "text": "x"},
	}, UpdateChatItemOptions{})
	if err == nil {
		t.Fatalf("expected invalid chatRef error")
	}
	if !strings.Contains(err.Error(), "invalid chatRef") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTextMessageInContact(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"chatItemNotChanged",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItem":{
					"chatInfo":{"type":"direct","contact":{"contactId":42}},
					"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"same"}}}
				}
			}
		}`)
	}()

	res, err := c.UpdateTextMessageInContact(ctx, 42, 7, "same", false)
	if err != nil {
		t.Fatalf("UpdateTextMessageInContact: %v", err)
	}
	cmd := <-done

	if res.ResponseType != types.ResponseTypeChatItemNotChanged {
		t.Fatalf("unexpected response type: %s", res.ResponseType)
	}
	if res.Updated {
		t.Fatalf("expected updated=false")
	}
	if !strings.HasPrefix(cmd, "/_update item @42 7") {
		t.Fatalf("unexpected contact update command: %q", cmd)
	}
}

func TestDeleteChatItems(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"chatItemsDeleted",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItemDeletions":[],
				"byUser":true,
				"timed":false
			}
		}`)
	}()

	if _, err := c.DeleteChatItems(ctx, "@42", []int64{1, 2}, CIDeleteModeBroadcast); err != nil {
		t.Fatalf("DeleteChatItems: %v", err)
	}
	cmd := <-done

	if !strings.HasPrefix(cmd, "/_delete item @42 1,2 broadcast") {
		t.Fatalf("unexpected delete command: %q", cmd)
	}
}

func TestDeleteChatItemsEmptyIDs(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	_, err = c.DeleteChatItems(context.Background(), "@42", nil, CIDeleteModeBroadcast)
	if err == nil {
		t.Fatalf("expected error for empty chatItemIDs")
	}
}

func TestDeleteChatItemsInvalidChatRef(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	_, err = c.DeleteChatItems(context.Background(), "bad-ref", []int64{1}, CIDeleteModeBroadcast)
	if err == nil {
		t.Fatalf("expected invalid chatRef error")
	}
	if !strings.Contains(err.Error(), "invalid chatRef") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModerateDeleteGroupChatItems(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"chatItemsDeleted",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItemDeletions":[],
				"byUser":true,
				"timed":false
			}
		}`)
	}()

	if _, err := c.ModerateDeleteGroupChatItems(ctx, 7, []int64{10, 11}); err != nil {
		t.Fatalf("ModerateDeleteGroupChatItems: %v", err)
	}
	cmd := <-done

	if !strings.HasPrefix(cmd, "/_delete member item #7 10,11") {
		t.Fatalf("unexpected moderate delete command: %q", cmd)
	}
}

func TestAddChatItemReaction(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"chatItemReaction",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"added":true,
				"reaction":{"kind":"like"}
			}
		}`)
	}()

	res, err := c.AddChatItemReaction(ctx, "@42", 9, map[string]any{"kind": "like"})
	if err != nil {
		t.Fatalf("AddChatItemReaction: %v", err)
	}
	cmd := <-done

	if !res.Added {
		t.Fatalf("expected reaction to be added")
	}
	if !strings.HasPrefix(cmd, "/_reaction @42 9 on") {
		t.Fatalf("unexpected add reaction command: %q", cmd)
	}
}

func TestRemoveChatItemReaction(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"chatItemReaction",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"added":false,
				"reaction":{"kind":"like"}
			}
		}`)
	}()

	res, err := c.RemoveChatItemReaction(ctx, "@42", 9, map[string]any{"kind": "like"})
	if err != nil {
		t.Fatalf("RemoveChatItemReaction: %v", err)
	}
	cmd := <-done

	if res.Added {
		t.Fatalf("expected reaction to be removed")
	}
	if !strings.HasPrefix(cmd, "/_reaction @42 9 off") {
		t.Fatalf("unexpected remove reaction command: %q", cmd)
	}
}

func TestSetChatItemReactionNilPayload(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	if _, err := c.SetChatItemReaction(context.Background(), "@42", 1, true, nil); err == nil {
		t.Fatalf("expected nil reaction error")
	}
}

func TestSetChatItemReactionInvalidChatRef(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	if _, err := c.SetChatItemReaction(context.Background(), "broken", 1, true, map[string]any{"kind": "like"}); err == nil {
		t.Fatalf("expected invalid chatRef error")
	}
}

func TestSetProfileAddress(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"userProfileUpdated",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"fromProfile":{"displayName":"bot"},
				"toProfile":{"displayName":"bot"},
				"updateSummary":{}
			}
		}`)
	}()

	if err := c.SetProfileAddress(ctx, 1, true); err != nil {
		t.Fatalf("SetProfileAddress: %v", err)
	}

	cmd := <-done
	if !strings.HasPrefix(cmd, "/_profile_address 1 on") {
		t.Fatalf("unexpected profile-address command: %q", cmd)
	}
}

func TestSetAddressSettings(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"userContactLinkUpdated",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"contactLink":{"connLinkContact":{"connFullLink":"smp://full","connShortLink":"smp://short"}}
			}
		}`)
	}()

	if err := c.SetAddressSettings(ctx, 1, map[string]any{"businessAddress": true}); err != nil {
		t.Fatalf("SetAddressSettings: %v", err)
	}

	cmd := <-done
	if !strings.HasPrefix(cmd, "/_address_settings 1 ") {
		t.Fatalf("unexpected address-settings command: %q", cmd)
	}
}

func TestDeleteChat(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"contactDeleted",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"contact":{"contactId":42,"profile":{"displayName":"alice"}}
			}
		}`)
	}()

	res, err := c.DeleteChat(ctx, "@42", ChatDeleteMode("entity"))
	if err != nil {
		t.Fatalf("DeleteChat: %v", err)
	}

	cmd := <-done
	if res.ResponseType != types.ResponseTypeContactDeleted {
		t.Fatalf("unexpected delete-chat response: %s", res.ResponseType)
	}
	if res.Contact == nil || res.Contact.ContactID != 42 {
		t.Fatalf("unexpected deleted contact payload: %+v", res.Contact)
	}
	if !strings.HasPrefix(cmd, "/_delete @42 entity") {
		t.Fatalf("unexpected delete-chat command: %q", cmd)
	}
}

func TestDeleteChatConnectionDeletedTypedPayload(t *testing.T) {
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
				"type":"contactConnectionDeleted",
				"user":{"userId":1},
				"connection":{"pccConnId":11,"pccAgentConnId":"a1","pccConnStatus":"connected","viaContactUri":true,"localAlias":"","createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:00:00Z"}
			}
		}`)
		close(done)
	}()

	res, err := c.DeleteChat(ctx, "@42", ChatDeleteMode("entity"))
	if err != nil {
		t.Fatalf("DeleteChat: %v", err)
	}
	<-done

	if res.ResponseType != types.ResponseTypeContactConnectionDeleted {
		t.Fatalf("unexpected delete-chat response: %s", res.ResponseType)
	}
	if string(res.Connection) == "" {
		t.Fatalf("expected raw connection payload")
	}
	if res.ConnectionInfo == nil || res.ConnectionInfo.PccConnId != 11 {
		t.Fatalf("unexpected typed connection payload: %+v", res.ConnectionInfo)
	}
}

func TestDeleteChatEmptyMode(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	if _, err := c.DeleteChat(context.Background(), "@42", ""); err == nil {
		t.Fatalf("expected empty mode error")
	}
}

func TestDeleteChatInvalidChatRef(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	if _, err := c.DeleteChat(context.Background(), "nope", ChatDeleteMode("entity")); err == nil {
		t.Fatalf("expected invalid chatRef error")
	}
}

func TestReceiveFileAccepted(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	c, err := New(transport)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer c.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	enc := true
	inline := false
	path := "/tmp/file.bin"

	done := make(chan string, 1)
	go func() {
		rawReq := <-transport.writeCh
		var req struct {
			CorrID string `json:"corrId"`
			Cmd    string `json:"cmd"`
		}
		_ = json.Unmarshal(rawReq, &req)
		done <- req.Cmd
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"rcvFileAccepted",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItem":{
					"chatInfo":{"type":"direct","contact":{"contactId":42}},
					"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"file"}}}
				}
			}
		}`)
	}()

	res, err := c.ReceiveFile(ctx, 99, ReceiveFileOptions{
		UserApprovedRelays: true,
		StoreEncrypted:     &enc,
		Inline:             &inline,
		Path:               &path,
	})
	if err != nil {
		t.Fatalf("ReceiveFile: %v", err)
	}

	cmd := <-done
	if res.ResponseType != types.ResponseTypeRcvFileAccepted {
		t.Fatalf("unexpected receive-file response: %s", res.ResponseType)
	}
	if res.ChatItem == nil {
		t.Fatalf("expected chat item payload")
	}
	if !strings.HasPrefix(cmd, "/freceive 99 approved_relays=on encrypt=on inline=off /tmp/file.bin") {
		t.Fatalf("unexpected receive-file command: %q", cmd)
	}
}

func TestReceiveFileAcceptedSndCancelled(t *testing.T) {
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
				"type":"rcvFileAcceptedSndCancelled",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"rcvFileTransfer":{"id":"rcv-1"}
			}
		}`)
		close(done)
	}()

	res, err := c.ReceiveFile(ctx, 100, ReceiveFileOptions{})
	if err != nil {
		t.Fatalf("ReceiveFile: %v", err)
	}
	<-done

	if res.ResponseType != types.ResponseTypeRcvFileAcceptedSndCancelled {
		t.Fatalf("unexpected receive-file response: %s", res.ResponseType)
	}
	if string(res.Transfer) == "" {
		t.Fatalf("expected transfer payload")
	}
	if res.TransferDetails == nil {
		t.Fatalf("expected typed transfer details")
	}
}

func TestCancelFileSndFileCancelled(t *testing.T) {
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
		transport.readCh <- []byte(`{
			"corrId":"` + req.CorrID + `",
			"resp":{
				"type":"sndFileCancelled",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItem_":{"chatInfo":{"type":"direct","contact":{"contactId":42}},"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"file"}}}},
				"fileTransferMeta":{"id":"snd-1"},
				"sndFileTransfers":[{"id":"s1"}]
			}
		}`)
	}()

	res, err := c.CancelFile(ctx, 77)
	if err != nil {
		t.Fatalf("CancelFile: %v", err)
	}
	cmd := <-done

	if res.ResponseType != types.ResponseTypeSndFileCancelled {
		t.Fatalf("unexpected cancel-file response: %s", res.ResponseType)
	}
	if res.ChatItem == nil {
		t.Fatalf("expected chat item payload")
	}
	if len(res.Transfers) != 1 {
		t.Fatalf("unexpected transfers count: %d", len(res.Transfers))
	}
	if res.TransferSnd == nil {
		t.Fatalf("expected typed snd transfer")
	}
	if len(res.TransfersSndDetail) != 1 {
		t.Fatalf("unexpected typed snd transfers count: %d", len(res.TransfersSndDetail))
	}
	if !strings.HasPrefix(cmd, "/fcancel 77") {
		t.Fatalf("unexpected cancel-file command: %q", cmd)
	}
}

func TestCancelFileRcvFileCancelled(t *testing.T) {
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
				"type":"rcvFileCancelled",
				"user":{"userId":1,"profile":{"displayName":"bot"}},
				"chatItem_":{"chatInfo":{"type":"direct","contact":{"contactId":42}},"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"file"}}}},
				"rcvFileTransfer":{"id":"rcv-1"}
			}
		}`)
		close(done)
	}()

	res, err := c.CancelFile(ctx, 78)
	if err != nil {
		t.Fatalf("CancelFile: %v", err)
	}
	<-done

	if res.ResponseType != types.ResponseTypeRcvFileCancelled {
		t.Fatalf("unexpected cancel-file response: %s", res.ResponseType)
	}
	if res.ChatItem == nil {
		t.Fatalf("expected chat item payload")
	}
	if string(res.Transfer) == "" {
		t.Fatalf("expected transfer payload")
	}
	if res.TransferRcv == nil {
		t.Fatalf("expected typed rcv transfer")
	}
}
