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
	payload := []map[string]any{
		map[string]any{
			"msgContent": map[string]any{
				"type": "text",
				"text": text,
			},
			"mentions": map[string]any{},
		},
	}

	result, err := c.SendAPISendMessages(ctx, command.APISendMessages{
		SendRef:          sendRef,
		LiveMessage:      false,
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
	return c.SendTextMessage(ctx, command.DirectRef(contactID), text)
}

func (c *Client) SendTextToGroup(ctx context.Context, groupID int64, text string) error {
	return c.SendTextMessage(ctx, command.GroupRef(groupID), text)
}
