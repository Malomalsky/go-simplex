package bot

import (
	"testing"

	"github.com/Malomalsky/go-simplex/sdk/protocol"
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
