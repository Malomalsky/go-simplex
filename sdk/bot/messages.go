package bot

import (
	"encoding/json"
	"fmt"

	"github.com/Malomalsky/go-simplex/sdk/protocol"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

type DirectTextMessage struct {
	ContactID int64
	Text      string
}

func ExtractDirectTextMessages(msg protocol.Message) ([]DirectTextMessage, error) {
	if msg.Resp.Type != string(types.EventTypeNewChatItems) {
		return nil, fmt.Errorf("unexpected event type: %s", msg.Resp.Type)
	}

	var payload struct {
		Type      string `json:"type"`
		ChatItems []struct {
			ChatInfo struct {
				Type    string `json:"type"`
				Contact struct {
					ContactID int64 `json:"contactId"`
				} `json:"contact"`
			} `json:"chatInfo"`
			ChatItem struct {
				Content struct {
					Type       string `json:"type"`
					MsgContent struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"msgContent"`
				} `json:"content"`
			} `json:"chatItem"`
		} `json:"chatItems"`
	}

	if err := json.Unmarshal(msg.Resp.Raw, &payload); err != nil {
		return nil, fmt.Errorf("decode newChatItems payload: %w", err)
	}

	out := make([]DirectTextMessage, 0, len(payload.ChatItems))
	for _, item := range payload.ChatItems {
		if item.ChatInfo.Type != "direct" {
			continue
		}
		if item.ChatItem.Content.Type != "rcvMsgContent" {
			continue
		}
		if item.ChatItem.Content.MsgContent.Type != "text" {
			continue
		}
		out = append(out, DirectTextMessage{
			ContactID: item.ChatInfo.Contact.ContactID,
			Text:      item.ChatItem.Content.MsgContent.Text,
		})
	}
	return out, nil
}
