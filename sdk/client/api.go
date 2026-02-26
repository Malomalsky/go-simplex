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
}

type SendTextOptions struct {
	Live bool
	TTL  *int64
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

func (c *Client) ListGroups(ctx context.Context, userID int64, contactID *int64, search string) ([]json.RawMessage, error) {
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
		return append([]json.RawMessage(nil), result.GroupsList.Groups...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
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

func (c *Client) ListGroupMembers(ctx context.Context, groupID int64) (json.RawMessage, error) {
	result, err := c.SendAPIListMembers(ctx, command.APIListMembers{GroupId: groupID})
	if err != nil {
		return nil, err
	}
	if result.GroupMembers != nil {
		return append([]byte(nil), result.GroupMembers.Group...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) CreateGroup(ctx context.Context, userID int64, incognito bool, groupProfile map[string]any) (json.RawMessage, error) {
	result, err := c.SendAPINewGroup(ctx, command.APINewGroup{
		UserId:       userID,
		Incognito:    incognito,
		GroupProfile: groupProfile,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupCreated != nil {
		return append([]byte(nil), result.GroupCreated.GroupInfo...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) UpdateGroupProfile(ctx context.Context, groupID int64, groupProfile map[string]any) (json.RawMessage, error) {
	result, err := c.SendAPIUpdateGroupProfile(ctx, command.APIUpdateGroupProfile{
		GroupId:      groupID,
		GroupProfile: groupProfile,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupUpdated != nil {
		return append([]byte(nil), result.GroupUpdated.ToGroup...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) CreateGroupLink(ctx context.Context, groupID int64, memberRole string) (json.RawMessage, error) {
	result, err := c.SendAPICreateGroupLink(ctx, command.APICreateGroupLink{
		GroupId:    groupID,
		MemberRole: memberRole,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupLinkCreated != nil {
		return append([]byte(nil), result.GroupLinkCreated.GroupLink...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}

func (c *Client) SetGroupLinkMemberRole(ctx context.Context, groupID int64, memberRole string) (json.RawMessage, error) {
	result, err := c.SendAPIGroupLinkMemberRole(ctx, command.APIGroupLinkMemberRole{
		GroupId:    groupID,
		MemberRole: memberRole,
	})
	if err != nil {
		return nil, err
	}
	if result.GroupLink != nil {
		return append([]byte(nil), result.GroupLink.GroupLink...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
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

func (c *Client) GetGroupLink(ctx context.Context, groupID int64) (json.RawMessage, error) {
	result, err := c.SendAPIGetGroupLink(ctx, command.APIGetGroupLink{GroupId: groupID})
	if err != nil {
		return nil, err
	}
	if result.GroupLink != nil {
		return append([]byte(nil), result.GroupLink.GroupLink...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
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

func (c *Client) ListUsers(ctx context.Context) ([]json.RawMessage, error) {
	result, err := c.SendListUsers(ctx, command.ListUsers{})
	if err != nil {
		return nil, err
	}
	if result.UsersList != nil {
		return append([]json.RawMessage(nil), result.UsersList.Users...), nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
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

func (c *Client) SendTextMessage(ctx context.Context, sendRef string, text string) error {
	return c.SendTextMessageWithOptions(ctx, sendRef, text, SendTextOptions{})
}

func (c *Client) SendTextMessageWithOptions(ctx context.Context, sendRef string, text string, options SendTextOptions) error {
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

func connectSummaryFromAPIConnectResult(result APIConnectResult) (*ConnectSummary, error) {
	if result.ContactAlreadyExists != nil {
		return &ConnectSummary{
			ResponseType:    types.ResponseTypeContactAlreadyExists,
			ExistingContact: &result.ContactAlreadyExists.Contact,
		}, nil
	}
	if result.SentConfirmation != nil {
		return &ConnectSummary{
			ResponseType: types.ResponseTypeSentConfirmation,
			Connection:   append([]byte(nil), result.SentConfirmation.Connection...),
		}, nil
	}
	if result.SentInvitation != nil {
		return &ConnectSummary{
			ResponseType: types.ResponseTypeSentInvitation,
			Connection:   append([]byte(nil), result.SentInvitation.Connection...),
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
		return &ConnectSummary{
			ResponseType: types.ResponseTypeSentConfirmation,
			Connection:   append([]byte(nil), result.SentConfirmation.Connection...),
		}, nil
	}
	if result.SentInvitation != nil {
		return &ConnectSummary{
			ResponseType: types.ResponseTypeSentInvitation,
			Connection:   append([]byte(nil), result.SentInvitation.Connection...),
		}, nil
	}
	if result.ChatCmdError != nil {
		return nil, commandErrorFromRaw(result.Message.Resp.Type, result.Message.Resp.Raw)
	}
	return nil, fmt.Errorf("missing response payload for %s", result.Message.Resp.Type)
}
