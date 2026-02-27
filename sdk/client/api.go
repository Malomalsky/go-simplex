package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Malomalsky/go-simplex/sdk/command"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

type CommandError struct {
	ResponseType string
	Payload      []byte
}

type ConnectSummary struct {
	ResponseType    types.ResponseType
	ExistingContact *types.Contact
	Connection      json.RawMessage
	ConnectionInfo  *types.PendingContactConnection
}

type SendTextOptions struct {
	Live bool
	TTL  *int64
}

type CIDeleteMode string

const (
	CIDeleteModeBroadcast    CIDeleteMode = "broadcast"
	CIDeleteModeInternal     CIDeleteMode = "internal"
	CIDeleteModeInternalMark CIDeleteMode = "internalMark"
)

type UpdateChatItemOptions struct {
	Live bool
}

type UpdateChatItemSummary struct {
	ResponseType types.ResponseType
	Updated      bool
	ChatItem     types.AChatItem
}

type ChatDeleteMode string

type DeleteChatSummary struct {
	ResponseType    types.ResponseType
	Contact         *types.Contact
	Connection      json.RawMessage
	ConnectionInfo  *types.PendingContactConnection
	GroupInfo       json.RawMessage
	GroupInfoRecord *types.GroupInfo
}

type ReceiveFileOptions struct {
	UserApprovedRelays bool
	StoreEncrypted     *bool
	Inline             *bool
	Path               *string
}

type ReceiveFileSummary struct {
	ResponseType    types.ResponseType
	ChatItem        *types.AChatItem
	Transfer        json.RawMessage
	TransferDetails *types.RcvFileTransfer
}

type CancelFileSummary struct {
	ResponseType       types.ResponseType
	ChatItem           *types.AChatItem
	Transfer           json.RawMessage
	TransferRcv        *types.RcvFileTransfer
	TransferSnd        *types.FileTransferMeta
	Transfers          []json.RawMessage
	TransfersSndDetail []types.SndFileTransfer
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("chat command error response: %s", e.ResponseType)
}

func commandErrorFromRaw(responseType string, raw []byte) *CommandError {
	return &CommandError{
		ResponseType: responseType,
		Payload:      append([]byte(nil), raw...),
	}
}

func (e *CommandError) IsStoreError(tag string) bool {
	if tag == "" || len(e.Payload) == 0 {
		return false
	}
	var payload struct {
		ChatError struct {
			Type       string `json:"type"`
			StoreError struct {
				Type string `json:"type"`
			} `json:"storeError"`
		} `json:"chatError"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return false
	}
	return payload.ChatError.Type == "errorStore" && payload.ChatError.StoreError.Type == tag
}

func (c *Client) GetActiveUser(ctx context.Context) (*types.User, error) {
	result, err := c.SendShowActiveUser(ctx, command.ShowActiveUser{})
	if err != nil {
		return nil, err
	}
	if result.ActiveUser != nil {
		return &result.ActiveUser.User, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) GetUserAddress(ctx context.Context, userID int64) (string, error) {
	result, err := c.SendAPIShowMyAddress(ctx, command.APIShowMyAddress{UserId: userID})
	if err != nil {
		return "", err
	}
	if result.UserContactLink != nil {
		return result.UserContactLink.ContactLink.ConnLinkContact.PreferredLink(), nil
	}
	if result.ChatCmdError != nil {
		return "", commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return "", fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) CreateUserAddress(ctx context.Context, userID int64) (string, error) {
	result, err := c.SendAPICreateMyAddress(ctx, command.APICreateMyAddress{UserId: userID})
	if err != nil {
		return "", err
	}
	if result.UserContactLinkCreated != nil {
		return result.UserContactLinkCreated.ConnLinkContact.PreferredLink(), nil
	}
	if result.ChatCmdError != nil {
		return "", commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return "", fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) DeleteUserAddress(ctx context.Context, userID int64) error {
	result, err := c.SendAPIDeleteMyAddress(ctx, command.APIDeleteMyAddress{UserId: userID})
	if err != nil {
		return err
	}
	if result.UserContactLinkDeleted != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetProfileAddress(ctx context.Context, userID int64, enable bool) error {
	result, err := c.SendAPISetProfileAddress(ctx, command.APISetProfileAddress{
		UserId: userID,
		Enable: enable,
	})
	if err != nil {
		return err
	}
	if result.UserProfileUpdated != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetAddressSettings(ctx context.Context, userID int64, settings map[string]any) error {
	result, err := c.SendAPISetAddressSettings(ctx, command.APISetAddressSettings{
		UserId:   userID,
		Settings: settings,
	})
	if err != nil {
		return err
	}
	if result.UserContactLinkUpdated != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) EnsureUserAddress(ctx context.Context, userID int64) (string, error) {
	addr, err := c.GetUserAddress(ctx, userID)
	if err == nil && addr != "" {
		return addr, nil
	}
	if err != nil {
		var cmdErr *CommandError
		if !errors.As(err, &cmdErr) || !cmdErr.IsStoreError("userContactLinkNotFound") {
			return "", err
		}
	}
	return c.CreateUserAddress(ctx, userID)
}

func (c *Client) ListContacts(ctx context.Context, userID int64) ([]types.Contact, error) {
	result, err := c.SendAPIListContacts(ctx, command.APIListContacts{UserId: userID})
	if err != nil {
		return nil, err
	}
	if result.ContactsList != nil {
		return append([]types.Contact(nil), result.ContactsList.Contacts...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ListGroupsTyped(ctx context.Context, userID int64, contactID *int64, search string) ([]types.GroupInfoSummary, error) {
	req := command.APIListGroups{
		UserId:     userID,
		ContactId_: contactID,
	}
	if search != "" {
		req.Search = &search
	}

	result, err := c.SendAPIListGroups(ctx, req)
	if err != nil {
		return nil, err
	}
	if result.GroupsList != nil {
		return append([]types.GroupInfoSummary(nil), result.GroupsList.Groups...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ListGroups(ctx context.Context, userID int64, contactID *int64, search string) ([]json.RawMessage, error) {
	groups, err := c.ListGroupsTyped(ctx, userID, contactID, search)
	if err != nil {
		return nil, err
	}
	return marshalRawSlice(groups)
}

func (c *Client) AddGroupMember(ctx context.Context, groupID int64, contactID int64, memberRole string) error {
	result, err := c.SendAPIAddMember(ctx, command.APIAddMember{
		GroupId:    groupID,
		ContactId:  contactID,
		MemberRole: memberRole,
	})
	if err != nil {
		return err
	}
	if result.SentGroupInvitation != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) JoinGroup(ctx context.Context, groupID int64) error {
	result, err := c.SendAPIJoinGroup(ctx, command.APIJoinGroup{GroupId: groupID})
	if err != nil {
		return err
	}
	if result.UserAcceptedGroupSent != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) AcceptGroupMember(ctx context.Context, groupID int64, groupMemberID int64, memberRole string) error {
	result, err := c.SendAPIAcceptMember(ctx, command.APIAcceptMember{
		GroupId:       groupID,
		GroupMemberId: groupMemberID,
		MemberRole:    memberRole,
	})
	if err != nil {
		return err
	}
	if result.MemberAccepted != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetGroupMembersRole(ctx context.Context, groupID int64, groupMemberIDs []int64, memberRole string) error {
	result, err := c.SendAPIMembersRole(ctx, command.APIMembersRole{
		GroupId:        groupID,
		GroupMemberIds: groupMemberIDs,
		MemberRole:     memberRole,
	})
	if err != nil {
		return err
	}
	if result.MembersRoleUser != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) BlockGroupMembersForAll(ctx context.Context, groupID int64, groupMemberIDs []int64, blocked bool) error {
	result, err := c.SendAPIBlockMembersForAll(ctx, command.APIBlockMembersForAll{
		GroupId:        groupID,
		GroupMemberIds: groupMemberIDs,
		Blocked:        blocked,
	})
	if err != nil {
		return err
	}
	if result.MembersBlockedForAllUser != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) RemoveGroupMembers(ctx context.Context, groupID int64, groupMemberIDs []int64, withMessages bool) error {
	result, err := c.SendAPIRemoveMembers(ctx, command.APIRemoveMembers{
		GroupId:        groupID,
		GroupMemberIds: groupMemberIDs,
		WithMessages:   withMessages,
	})
	if err != nil {
		return err
	}
	if result.UserDeletedMembers != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) LeaveGroup(ctx context.Context, groupID int64) error {
	result, err := c.SendAPILeaveGroup(ctx, command.APILeaveGroup{GroupId: groupID})
	if err != nil {
		return err
	}
	if result.LeftMemberUser != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ListGroupMembersTyped(ctx context.Context, groupID int64) (*types.Group, error) {
	result, err := c.SendAPIListMembers(ctx, command.APIListMembers{GroupId: groupID})
	if err != nil {
		return nil, err
	}
	if result.GroupMembers != nil {
		group := result.GroupMembers.Group
		return &group, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ListGroupMembers(ctx context.Context, groupID int64) (json.RawMessage, error) {
	group, err := c.ListGroupMembersTyped(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return marshalRaw(group)
}

func (c *Client) CreateGroupTyped(ctx context.Context, userID int64, incognito bool, groupProfile map[string]any) (*types.GroupInfo, error) {
	result, err := c.SendAPINewGroup(ctx, command.APINewGroup{
		UserId:       userID,
		Incognito:    incognito,
		GroupProfile: groupProfile,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupCreated != nil {
		groupInfo := result.GroupCreated.GroupInfo
		return &groupInfo, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) CreateGroup(ctx context.Context, userID int64, incognito bool, groupProfile map[string]any) (json.RawMessage, error) {
	groupInfo, err := c.CreateGroupTyped(ctx, userID, incognito, groupProfile)
	if err != nil {
		return nil, err
	}
	return marshalRaw(groupInfo)
}

func (c *Client) UpdateGroupProfileTyped(ctx context.Context, groupID int64, groupProfile map[string]any) (*types.GroupInfo, error) {
	result, err := c.SendAPIUpdateGroupProfile(ctx, command.APIUpdateGroupProfile{
		GroupId:      groupID,
		GroupProfile: groupProfile,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupUpdated != nil {
		groupInfo := result.GroupUpdated.ToGroup
		return &groupInfo, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) UpdateGroupProfile(ctx context.Context, groupID int64, groupProfile map[string]any) (json.RawMessage, error) {
	groupInfo, err := c.UpdateGroupProfileTyped(ctx, groupID, groupProfile)
	if err != nil {
		return nil, err
	}
	return marshalRaw(groupInfo)
}

func (c *Client) CreateGroupLinkTyped(ctx context.Context, groupID int64, memberRole string) (*types.GroupLink, error) {
	result, err := c.SendAPICreateGroupLink(ctx, command.APICreateGroupLink{
		GroupId:    groupID,
		MemberRole: memberRole,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupLinkCreated != nil {
		groupLink := result.GroupLinkCreated.GroupLink
		return &groupLink, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) CreateGroupLink(ctx context.Context, groupID int64, memberRole string) (json.RawMessage, error) {
	groupLink, err := c.CreateGroupLinkTyped(ctx, groupID, memberRole)
	if err != nil {
		return nil, err
	}
	return marshalRaw(groupLink)
}

func (c *Client) SetGroupLinkMemberRoleTyped(ctx context.Context, groupID int64, memberRole string) (*types.GroupLink, error) {
	result, err := c.SendAPIGroupLinkMemberRole(ctx, command.APIGroupLinkMemberRole{
		GroupId:    groupID,
		MemberRole: memberRole,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupLink != nil {
		groupLink := result.GroupLink.GroupLink
		return &groupLink, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetGroupLinkMemberRole(ctx context.Context, groupID int64, memberRole string) (json.RawMessage, error) {
	groupLink, err := c.SetGroupLinkMemberRoleTyped(ctx, groupID, memberRole)
	if err != nil {
		return nil, err
	}
	return marshalRaw(groupLink)
}

func (c *Client) DeleteGroupLink(ctx context.Context, groupID int64) error {
	result, err := c.SendAPIDeleteGroupLink(ctx, command.APIDeleteGroupLink{GroupId: groupID})
	if err != nil {
		return err
	}
	if result.GroupLinkDeleted != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) GetGroupLinkTyped(ctx context.Context, groupID int64) (*types.GroupLink, error) {
	result, err := c.SendAPIGetGroupLink(ctx, command.APIGetGroupLink{GroupId: groupID})
	if err != nil {
		return nil, err
	}
	if result.GroupLink != nil {
		groupLink := result.GroupLink.GroupLink
		return &groupLink, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) GetGroupLink(ctx context.Context, groupID int64) (json.RawMessage, error) {
	groupLink, err := c.GetGroupLinkTyped(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return marshalRaw(groupLink)
}

func (c *Client) CreateContactInvitation(ctx context.Context, userID int64, incognito bool) (string, error) {
	result, err := c.SendAPIAddContact(ctx, command.APIAddContact{
		UserId:    userID,
		Incognito: incognito,
	})
	if err != nil {
		return "", err
	}
	if result.Invitation != nil {
		return result.Invitation.ConnLinkInvitation.PreferredLink(), nil
	}
	if result.ChatCmdError != nil {
		return "", commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return "", fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ConnectPlan(ctx context.Context, userID int64, connectionLink string) (*types.ResponseConnectionPlan, error) {
	var link *string
	if connectionLink != "" {
		link = &connectionLink
	}

	result, err := c.SendAPIConnectPlan(ctx, command.APIConnectPlan{
		UserId:         userID,
		ConnectionLink: link,
	})
	if err != nil {
		return nil, err
	}
	if result.ConnectionPlan != nil {
		return result.ConnectionPlan, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ConnectWithPreparedLink(ctx context.Context, userID int64, incognito bool, preparedLink *string) (*ConnectSummary, error) {
	result, err := c.SendAPIConnect(ctx, command.APIConnect{
		UserId:        userID,
		Incognito:     incognito,
		PreparedLink_: preparedLink,
	})
	if err != nil {
		return nil, err
	}
	return connectSummaryFromAPIConnectResult(result)
}

func (c *Client) ConnectWithLink(ctx context.Context, connLink *string) (*ConnectSummary, error) {
	result, err := c.SendConnect(ctx, command.Connect{
		Incognito: false,
		ConnLink_: connLink,
	})
	if err != nil {
		return nil, err
	}
	return connectSummaryFromConnectResult(result)
}

func (c *Client) CreateUser(ctx context.Context, newUser map[string]any) (*types.User, error) {
	result, err := c.SendCreateActiveUser(ctx, command.CreateActiveUser{NewUser: newUser})
	if err != nil {
		return nil, err
	}
	if result.ActiveUser != nil {
		return &result.ActiveUser.User, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ListUsersTyped(ctx context.Context) ([]types.UserInfo, error) {
	result, err := c.SendListUsers(ctx, command.ListUsers{})
	if err != nil {
		return nil, err
	}
	if result.UsersList != nil {
		return append([]types.UserInfo(nil), result.UsersList.Users...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) ListUsers(ctx context.Context) ([]json.RawMessage, error) {
	users, err := c.ListUsersTyped(ctx)
	if err != nil {
		return nil, err
	}
	return marshalRawSlice(users)
}

func (c *Client) SetActiveUser(ctx context.Context, userID int64, viewPwd *string) (*types.User, error) {
	result, err := c.SendAPISetActiveUser(ctx, command.APISetActiveUser{
		UserId:  userID,
		ViewPwd: viewPwd,
	})
	if err != nil {
		return nil, err
	}
	if result.ActiveUser != nil {
		return &result.ActiveUser.User, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) DeleteUser(ctx context.Context, userID int64, delSMPQueues bool, viewPwd *string) error {
	result, err := c.SendAPIDeleteUser(ctx, command.APIDeleteUser{
		UserId:       userID,
		DelSMPQueues: delSMPQueues,
		ViewPwd:      viewPwd,
	})
	if err != nil {
		return err
	}
	if result.CmdOk != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) UpdateProfile(ctx context.Context, userID int64, profile map[string]any) (bool, error) {
	result, err := c.SendAPIUpdateProfile(ctx, command.APIUpdateProfile{
		UserId:  userID,
		Profile: profile,
	})
	if err != nil {
		return false, err
	}
	if result.UserProfileUpdated != nil {
		return true, nil
	}
	if result.UserProfileNoChange != nil {
		return false, nil
	}
	if result.ChatCmdError != nil {
		return false, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return false, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetContactPreferences(ctx context.Context, contactID int64, preferences map[string]any) error {
	result, err := c.SendAPISetContactPrefs(ctx, command.APISetContactPrefs{
		ContactId:   contactID,
		Preferences: preferences,
	})
	if err != nil {
		return err
	}
	if result.ContactPrefsUpdated != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) AcceptContactRequest(ctx context.Context, contactReqID int64) error {
	result, err := c.SendAPIAcceptContact(ctx, command.APIAcceptContact{ContactReqId: contactReqID})
	if err != nil {
		return err
	}
	if result.AcceptingContactRequest != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) RejectContactRequest(ctx context.Context, contactReqID int64) error {
	result, err := c.SendAPIRejectContact(ctx, command.APIRejectContact{ContactReqId: contactReqID})
	if err != nil {
		return err
	}
	if result.ContactRequestRejected != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) EnableAddressAutoAccept(ctx context.Context, userID int64) error {
	settings := map[string]any{
		"businessAddress": false,
		"autoAccept": map[string]any{
			"acceptIncognito": false,
		},
	}
	return c.SetAddressSettings(ctx, userID, settings)
}

func (c *Client) SendTextMessage(ctx context.Context, sendRef string, text string) error {
	return c.SendTextMessageWithOptions(ctx, sendRef, text, SendTextOptions{})
}

func (c *Client) SendTextMessageWithOptions(ctx context.Context, sendRef string, text string, options SendTextOptions) error {
	if err := command.ValidateRef(sendRef); err != nil {
		return fmt.Errorf("invalid sendRef: %w", err)
	}

	payload := []map[string]any{
		{
			"msgContent": map[string]any{
				"type": "text",
				"text": text,
			},
			"mentions": map[string]any{},
		},
	}

	result, err := c.SendAPISendMessages(ctx, command.APISendMessages{
		SendRef:          sendRef,
		LiveMessage:      options.Live,
		Ttl:              options.TTL,
		ComposedMessages: payload,
	})
	if err != nil {
		return err
	}
	if result.NewChatItems != nil {
		return nil
	}
	if result.ChatCmdError != nil {
		return commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SendTextToContact(ctx context.Context, contactID int64, text string) error {
	return c.SendTextToContactWithOptions(ctx, contactID, text, SendTextOptions{})
}

func (c *Client) SendTextToContactWithOptions(ctx context.Context, contactID int64, text string, options SendTextOptions) error {
	return c.SendTextMessageWithOptions(ctx, command.DirectRef(contactID), text, options)
}

func (c *Client) SendTextToGroup(ctx context.Context, groupID int64, text string) error {
	return c.SendTextToGroupWithOptions(ctx, groupID, text, SendTextOptions{})
}

func (c *Client) SendTextToGroupWithOptions(ctx context.Context, groupID int64, text string, options SendTextOptions) error {
	return c.SendTextMessageWithOptions(ctx, command.GroupRef(groupID), text, options)
}

func (c *Client) UpdateChatItem(ctx context.Context, chatRef string, chatItemID int64, updatedMessage map[string]any, options UpdateChatItemOptions) (*UpdateChatItemSummary, error) {
	if err := command.ValidateRef(chatRef); err != nil {
		return nil, fmt.Errorf("invalid chatRef: %w", err)
	}

	result, err := c.SendAPIUpdateChatItem(ctx, command.APIUpdateChatItem{
		ChatRef:        chatRef,
		ChatItemId:     chatItemID,
		LiveMessage:    options.Live,
		UpdatedMessage: updatedMessage,
	})
	if err != nil {
		return nil, err
	}
	if result.ChatItemUpdated != nil {
		return &UpdateChatItemSummary{
			ResponseType: types.ResponseTypeChatItemUpdated,
			Updated:      true,
			ChatItem:     result.ChatItemUpdated.ChatItem,
		}, nil
	}
	if result.ChatItemNotChanged != nil {
		return &UpdateChatItemSummary{
			ResponseType: types.ResponseTypeChatItemNotChanged,
			Updated:      false,
			ChatItem:     result.ChatItemNotChanged.ChatItem,
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) UpdateTextMessage(ctx context.Context, chatRef string, chatItemID int64, text string, live bool) (*UpdateChatItemSummary, error) {
	return c.UpdateChatItem(ctx, chatRef, chatItemID, map[string]any{
		"msgContent": map[string]any{
			"type": "text",
			"text": text,
		},
	}, UpdateChatItemOptions{Live: live})
}

func (c *Client) UpdateTextMessageInContact(ctx context.Context, contactID int64, chatItemID int64, text string, live bool) (*UpdateChatItemSummary, error) {
	return c.UpdateTextMessage(ctx, command.DirectRef(contactID), chatItemID, text, live)
}

func (c *Client) UpdateTextMessageInGroup(ctx context.Context, groupID int64, chatItemID int64, text string, live bool) (*UpdateChatItemSummary, error) {
	return c.UpdateTextMessage(ctx, command.GroupRef(groupID), chatItemID, text, live)
}

func (c *Client) DeleteChatItems(ctx context.Context, chatRef string, chatItemIDs []int64, deleteMode CIDeleteMode) (*types.ResponseChatItemsDeleted, error) {
	if err := command.ValidateRef(chatRef); err != nil {
		return nil, fmt.Errorf("invalid chatRef: %w", err)
	}
	if len(chatItemIDs) == 0 {
		return nil, fmt.Errorf("chatItemIDs is empty")
	}
	mode := deleteMode
	if mode == "" {
		mode = CIDeleteModeBroadcast
	}

	result, err := c.SendAPIDeleteChatItem(ctx, command.APIDeleteChatItem{
		ChatRef:     chatRef,
		ChatItemIds: append([]int64(nil), chatItemIDs...),
		DeleteMode:  string(mode),
	})
	if err != nil {
		return nil, err
	}
	if result.ChatItemsDeleted != nil {
		return result.ChatItemsDeleted, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) DeleteChatItemsInContact(ctx context.Context, contactID int64, chatItemIDs []int64, deleteMode CIDeleteMode) (*types.ResponseChatItemsDeleted, error) {
	return c.DeleteChatItems(ctx, command.DirectRef(contactID), chatItemIDs, deleteMode)
}

func (c *Client) DeleteChatItemsInGroup(ctx context.Context, groupID int64, chatItemIDs []int64, deleteMode CIDeleteMode) (*types.ResponseChatItemsDeleted, error) {
	return c.DeleteChatItems(ctx, command.GroupRef(groupID), chatItemIDs, deleteMode)
}

func (c *Client) ModerateDeleteGroupChatItems(ctx context.Context, groupID int64, chatItemIDs []int64) (*types.ResponseChatItemsDeleted, error) {
	if len(chatItemIDs) == 0 {
		return nil, fmt.Errorf("chatItemIDs is empty")
	}

	result, err := c.SendAPIDeleteMemberChatItem(ctx, command.APIDeleteMemberChatItem{
		GroupId:     groupID,
		ChatItemIds: append([]int64(nil), chatItemIDs...),
	})
	if err != nil {
		return nil, err
	}
	if result.ChatItemsDeleted != nil {
		return result.ChatItemsDeleted, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetChatItemReaction(ctx context.Context, chatRef string, chatItemID int64, add bool, reaction map[string]any) (*types.ResponseChatItemReaction, error) {
	if err := command.ValidateRef(chatRef); err != nil {
		return nil, fmt.Errorf("invalid chatRef: %w", err)
	}
	if reaction == nil {
		return nil, fmt.Errorf("reaction is nil")
	}

	result, err := c.SendAPIChatItemReaction(ctx, command.APIChatItemReaction{
		ChatRef:    chatRef,
		ChatItemId: chatItemID,
		Add:        add,
		Reaction:   reaction,
	})
	if err != nil {
		return nil, err
	}
	if result.ChatItemReaction != nil {
		return result.ChatItemReaction, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) AddChatItemReaction(ctx context.Context, chatRef string, chatItemID int64, reaction map[string]any) (*types.ResponseChatItemReaction, error) {
	return c.SetChatItemReaction(ctx, chatRef, chatItemID, true, reaction)
}

func (c *Client) RemoveChatItemReaction(ctx context.Context, chatRef string, chatItemID int64, reaction map[string]any) (*types.ResponseChatItemReaction, error) {
	return c.SetChatItemReaction(ctx, chatRef, chatItemID, false, reaction)
}

func (c *Client) DeleteChat(ctx context.Context, chatRef string, mode ChatDeleteMode) (*DeleteChatSummary, error) {
	if err := command.ValidateRef(chatRef); err != nil {
		return nil, fmt.Errorf("invalid chatRef: %w", err)
	}
	if mode == "" {
		return nil, fmt.Errorf("chat delete mode is empty")
	}

	result, err := c.SendAPIDeleteChat(ctx, command.APIDeleteChat{
		ChatRef:        chatRef,
		ChatDeleteMode: string(mode),
	})
	if err != nil {
		return nil, err
	}
	if result.ContactDeleted != nil {
		contact := result.ContactDeleted.Contact
		return &DeleteChatSummary{
			ResponseType: types.ResponseTypeContactDeleted,
			Contact:      &contact,
		}, nil
	}
	if result.ContactConnectionDeleted != nil {
		connectionRaw, err := marshalRaw(result.ContactConnectionDeleted.Connection)
		if err != nil {
			return nil, err
		}
		connection := result.ContactConnectionDeleted.Connection
		return &DeleteChatSummary{
			ResponseType:   types.ResponseTypeContactConnectionDeleted,
			Connection:     connectionRaw,
			ConnectionInfo: &connection,
		}, nil
	}
	if result.GroupDeletedUser != nil {
		groupInfoRaw, err := marshalRaw(result.GroupDeletedUser.GroupInfo)
		if err != nil {
			return nil, err
		}
		groupInfo := result.GroupDeletedUser.GroupInfo
		return &DeleteChatSummary{
			ResponseType:    types.ResponseTypeGroupDeletedUser,
			GroupInfo:       groupInfoRaw,
			GroupInfoRecord: &groupInfo,
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) DeleteContactChat(ctx context.Context, contactID int64, mode ChatDeleteMode) (*DeleteChatSummary, error) {
	return c.DeleteChat(ctx, command.DirectRef(contactID), mode)
}

func (c *Client) DeleteGroupChat(ctx context.Context, groupID int64, mode ChatDeleteMode) (*DeleteChatSummary, error) {
	return c.DeleteChat(ctx, command.GroupRef(groupID), mode)
}

func (c *Client) ReceiveFile(ctx context.Context, fileID int64, options ReceiveFileOptions) (*ReceiveFileSummary, error) {
	result, err := c.SendReceiveFile(ctx, command.ReceiveFile{
		FileId:             fileID,
		UserApprovedRelays: options.UserApprovedRelays,
		StoreEncrypted:     options.StoreEncrypted,
		FileInline:         options.Inline,
		FilePath:           options.Path,
	})
	if err != nil {
		return nil, err
	}
	if result.RcvFileAccepted != nil {
		chatItem := result.RcvFileAccepted.ChatItem
		return &ReceiveFileSummary{
			ResponseType: types.ResponseTypeRcvFileAccepted,
			ChatItem:     &chatItem,
		}, nil
	}
	if result.RcvFileAcceptedSndCancelled != nil {
		transferRaw, err := marshalRaw(result.RcvFileAcceptedSndCancelled.RcvFileTransfer)
		if err != nil {
			return nil, err
		}
		transfer := result.RcvFileAcceptedSndCancelled.RcvFileTransfer
		return &ReceiveFileSummary{
			ResponseType:    types.ResponseTypeRcvFileAcceptedSndCancelled,
			Transfer:        transferRaw,
			TransferDetails: &transfer,
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) CancelFile(ctx context.Context, fileID int64) (*CancelFileSummary, error) {
	result, err := c.SendCancelFile(ctx, command.CancelFile{FileId: fileID})
	if err != nil {
		return nil, err
	}
	if result.SndFileCancelled != nil {
		var chatItem *types.AChatItem
		if result.SndFileCancelled.ChatItem_ != nil {
			item := *result.SndFileCancelled.ChatItem_
			chatItem = &item
		}
		transferRaw, err := marshalRaw(result.SndFileCancelled.FileTransferMeta)
		if err != nil {
			return nil, err
		}
		transfersRaw, err := marshalRawSlice(result.SndFileCancelled.SndFileTransfers)
		if err != nil {
			return nil, err
		}
		return &CancelFileSummary{
			ResponseType:       types.ResponseTypeSndFileCancelled,
			ChatItem:           chatItem,
			Transfer:           transferRaw,
			TransferSnd:        &result.SndFileCancelled.FileTransferMeta,
			Transfers:          transfersRaw,
			TransfersSndDetail: append([]types.SndFileTransfer(nil), result.SndFileCancelled.SndFileTransfers...),
		}, nil
	}
	if result.RcvFileCancelled != nil {
		var chatItem *types.AChatItem
		if result.RcvFileCancelled.ChatItem_ != nil {
			item := *result.RcvFileCancelled.ChatItem_
			chatItem = &item
		}
		transferRaw, err := marshalRaw(result.RcvFileCancelled.RcvFileTransfer)
		if err != nil {
			return nil, err
		}
		transfer := result.RcvFileCancelled.RcvFileTransfer
		return &CancelFileSummary{
			ResponseType: types.ResponseTypeRcvFileCancelled,
			ChatItem:     chatItem,
			Transfer:     transferRaw,
			TransferRcv:  &transfer,
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func marshalRaw(v any) (json.RawMessage, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal response payload: %w", err)
	}
	return raw, nil
}

func marshalRawSlice[T any](items []T) ([]json.RawMessage, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		raw, err := marshalRaw(item)
		if err != nil {
			return nil, err
		}
		out = append(out, raw)
	}
	return out, nil
}

func connectSummaryFromAPIConnectResult(result APIConnectResult) (*ConnectSummary, error) {
	if result.ContactAlreadyExists != nil {
		return &ConnectSummary{
			ResponseType:    types.ResponseTypeContactAlreadyExists,
			ExistingContact: &result.ContactAlreadyExists.Contact,
		}, nil
	}
	if result.SentConfirmation != nil {
		connectionRaw, err := marshalRaw(result.SentConfirmation.Connection)
		if err != nil {
			return nil, err
		}
		connection := result.SentConfirmation.Connection
		return &ConnectSummary{
			ResponseType:   types.ResponseTypeSentConfirmation,
			Connection:     connectionRaw,
			ConnectionInfo: &connection,
		}, nil
	}
	if result.SentInvitation != nil {
		connectionRaw, err := marshalRaw(result.SentInvitation.Connection)
		if err != nil {
			return nil, err
		}
		connection := result.SentInvitation.Connection
		return &ConnectSummary{
			ResponseType:   types.ResponseTypeSentInvitation,
			Connection:     connectionRaw,
			ConnectionInfo: &connection,
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func connectSummaryFromConnectResult(result ConnectResult) (*ConnectSummary, error) {
	if result.ContactAlreadyExists != nil {
		return &ConnectSummary{
			ResponseType:    types.ResponseTypeContactAlreadyExists,
			ExistingContact: &result.ContactAlreadyExists.Contact,
		}, nil
	}
	if result.SentConfirmation != nil {
		connectionRaw, err := marshalRaw(result.SentConfirmation.Connection)
		if err != nil {
			return nil, err
		}
		connection := result.SentConfirmation.Connection
		return &ConnectSummary{
			ResponseType:   types.ResponseTypeSentConfirmation,
			Connection:     connectionRaw,
			ConnectionInfo: &connection,
		}, nil
	}
	if result.SentInvitation != nil {
		connectionRaw, err := marshalRaw(result.SentInvitation.Connection)
		if err != nil {
			return nil, err
		}
		connection := result.SentInvitation.Connection
		return &ConnectSummary{
			ResponseType:   types.ResponseTypeSentInvitation,
			Connection:     connectionRaw,
			ConnectionInfo: &connection,
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}
