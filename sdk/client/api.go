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
	msg, err := c.Send(ctx, command.ShowActiveUser{})
	if err != nil {
		return nil, err
	}

	switch msg.Resp.Type {
	case "activeUser":
		var resp types.ActiveUserResp
		if err := msg.Resp.Decode(&resp); err != nil {
			return nil, err
		}
		return &resp.User, nil
	case "chatCmdError":
		return nil, &CommandError{ResponseType: msg.Resp.Type, Payload: append([]byte(nil), msg.Resp.Raw...)}
	default:
		return nil, fmt.Errorf("unexpected response type: %s", msg.Resp.Type)
	}
}

func (c *Client) GetUserAddress(ctx context.Context, userID int64) (string, error) {
	msg, err := c.Send(ctx, command.APIShowMyAddress{UserID: userID})
	if err != nil {
		return "", err
	}

	switch msg.Resp.Type {
	case "userContactLink":
		var resp types.UserContactLinkResp
		if err := msg.Resp.Decode(&resp); err != nil {
			return "", err
		}
		return resp.ContactLink.ConnLinkContact.PreferredLink(), nil
	case "chatCmdError":
		return "", &CommandError{ResponseType: msg.Resp.Type, Payload: append([]byte(nil), msg.Resp.Raw...)}
	default:
		return "", fmt.Errorf("unexpected response type: %s", msg.Resp.Type)
	}
}

func (c *Client) CreateUserAddress(ctx context.Context, userID int64) (string, error) {
	msg, err := c.Send(ctx, command.APICreateMyAddress{UserID: userID})
	if err != nil {
		return "", err
	}

	switch msg.Resp.Type {
	case "userContactLinkCreated":
		var resp types.UserContactLinkCreatedResp
		if err := msg.Resp.Decode(&resp); err != nil {
			return "", err
		}
		return resp.ConnLinkContact.PreferredLink(), nil
	case "chatCmdError":
		return "", &CommandError{ResponseType: msg.Resp.Type, Payload: append([]byte(nil), msg.Resp.Raw...)}
	default:
		return "", fmt.Errorf("unexpected response type: %s", msg.Resp.Type)
	}
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

func (c *Client) EnableAddressAutoAccept(ctx context.Context, userID int64) error {
	settings := map[string]any{
		"businessAddress": false,
		"autoAccept": map[string]any{
			"acceptIncognito": false,
		},
	}

	msg, err := c.Send(ctx, command.APISetAddressSettings{
		UserID:   userID,
		Settings: settings,
	})
	if err != nil {
		return err
	}

	switch msg.Resp.Type {
	case "userContactLinkUpdated":
		return nil
	case "chatCmdError":
		return &CommandError{ResponseType: msg.Resp.Type, Payload: append([]byte(nil), msg.Resp.Raw...)}
	default:
		return fmt.Errorf("unexpected response type: %s", msg.Resp.Type)
	}
}

func (c *Client) SendTextMessage(ctx context.Context, sendRef string, text string) error {
	payload := []any{
		map[string]any{
			"msgContent": map[string]any{
				"type": "text",
				"text": text,
			},
			"mentions": map[string]any{},
		},
	}

	msg, err := c.Send(ctx, command.APISendMessages{
		SendRef:          sendRef,
		LiveMessage:      false,
		ComposedMessages: payload,
	})
	if err != nil {
		return err
	}

	switch msg.Resp.Type {
	case "newChatItems":
		return nil
	case "chatCmdError":
		return &CommandError{ResponseType: msg.Resp.Type, Payload: append([]byte(nil), msg.Resp.Raw...)}
	default:
		return fmt.Errorf("unexpected response type: %s", msg.Resp.Type)
	}
}
