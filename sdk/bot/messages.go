package bot

import (
	"context"
	"fmt"

	"github.com/Malomalsky/go-simplex/sdk/client"
	"github.com/Malomalsky/go-simplex/sdk/protocol"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

type DirectTextMessage struct {
	ContactID int64
	Text      string
}

func (m DirectTextMessage) Reply(ctx context.Context, cli *client.Client, text string) error {
	if cli == nil {
		return fmt.Errorf("client is nil")
	}
	return cli.SendTextToContact(ctx, m.ContactID, text)
}

func ExtractDirectTextMessages(msg protocol.Message) ([]DirectTextMessage, error) {
	if msg.Resp.Type != string(types.EventTypeNewChatItems) {
		return nil, fmt.Errorf("unexpected event type: %s", msg.Resp.Type)
	}

	var payload types.EventNewChatItems
	if err := msg.Resp.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode newChatItems payload: %w", err)
	}

	return ExtractDirectTextMessagesFromNewChatItems(payload), nil
}

func ExtractDirectTextMessagesFromNewChatItems(payload types.EventNewChatItems) []DirectTextMessage {
	out := make([]DirectTextMessage, 0, len(payload.ChatItems))
	for _, item := range payload.ChatItems {
		if item.ChatInfo.Type != "direct" {
			continue
		}
		if item.ChatInfo.Contact == nil {
			continue
		}
		if item.ChatItem.Content.Type != "rcvMsgContent" {
			continue
		}
		if item.ChatItem.Content.MsgContent == nil || item.ChatItem.Content.MsgContent.Type != "text" {
			continue
		}
		out = append(out, DirectTextMessage{
			ContactID: item.ChatInfo.Contact.ContactID,
			Text:      item.ChatItem.Content.MsgContent.Text,
		})
	}
	return out
}
