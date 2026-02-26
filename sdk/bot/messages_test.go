package bot

import (
	"testing"

	"github.com/Malomalsky/go-simplex/sdk/protocol"
	"github.com/Malomalsky/go-simplex/sdk/types"
)

func TestExtractDirectTextMessages(t *testing.T) {
	t.Parallel()

	msg := protocol.Message{
		Resp: protocol.RawResponse{
			Type: "newChatItems",
			Raw: []byte(`{
				"type":"newChatItems",
				"chatItems":[
					{
						"chatInfo":{"type":"direct","contact":{"contactId":42}},
						"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"hello"}}}
					},
					{
						"chatInfo":{"type":"group"},
						"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"ignored"}}}
					}
				]
			}`),
		},
	}

	got, err := ExtractDirectTextMessages(msg)
	if err != nil {
		t.Fatalf("extract messages: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 message, got %d", len(got))
	}
	if got[0].ContactID != 42 || got[0].Text != "hello" {
		t.Fatalf("unexpected message: %+v", got[0])
	}
}

func TestExtractDirectTextMessagesFromNewChatItems(t *testing.T) {
	t.Parallel()

	payload := types.EventNewChatItems{
		ChatItems: []types.AChatItem{
			{
				ChatInfo: types.ChatInfo{
					Type: "direct",
					Contact: &types.Contact{
						ContactID: 7,
					},
				},
				ChatItem: types.ChatItem{
					Content: types.ChatContent{
						Type: "rcvMsgContent",
						MsgContent: &types.MsgContent{
							Type: "text",
							Text: "ping",
						},
					},
				},
			},
			{
				ChatInfo: types.ChatInfo{Type: "group"},
				ChatItem: types.ChatItem{
					Content: types.ChatContent{
						Type: "rcvMsgContent",
						MsgContent: &types.MsgContent{
							Type: "text",
							Text: "ignored",
						},
					},
				},
			},
		},
	}

	got := ExtractDirectTextMessagesFromNewChatItems(payload)
	if len(got) != 1 {
		t.Fatalf("expected 1 message, got %d", len(got))
	}
	if got[0].ContactID != 7 || got[0].Text != "ping" {
		t.Fatalf("unexpected message: %+v", got[0])
	}
}
