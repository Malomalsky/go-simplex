package types

import "testing"

func TestDecodeResponseByType(t *testing.T) {
	t.Parallel()

	v, err := DecodeResponseByType(ResponseTypeActiveUser, []byte(`{"type":"activeUser","user":{"userId":123,"profile":{"displayName":"bot"}}}`))
	if err != nil {
		t.Fatalf("decode response by type: %v", err)
	}

	resp, ok := v.(ResponseActiveUser)
	if !ok {
		t.Fatalf("unexpected response type: %T", v)
	}
	if resp.User.UserID != 123 {
		t.Fatalf("unexpected user id: %d", resp.User.UserID)
	}
}

func TestDecodeEventByType(t *testing.T) {
	t.Parallel()

	v, err := DecodeEventByType(EventTypeNewChatItems, []byte(`{
		"type":"newChatItems",
		"user":{"userId":1,"profile":{"displayName":"bot"}},
		"chatItems":[
			{"chatInfo":{"type":"direct","contact":{"contactId":42,"profile":{"displayName":"u"}}},"chatItem":{"content":{"type":"rcvMsgContent","msgContent":{"type":"text","text":"hello"}}}}
		]
	}`))
	if err != nil {
		t.Fatalf("decode event by type: %v", err)
	}

	evt, ok := v.(EventNewChatItems)
	if !ok {
		t.Fatalf("unexpected event type: %T", v)
	}
	if len(evt.ChatItems) != 1 {
		t.Fatalf("unexpected chat item count: %d", len(evt.ChatItems))
	}
	if evt.ChatItems[0].ChatInfo.Contact == nil || evt.ChatItems[0].ChatInfo.Contact.ContactID != 42 {
		t.Fatalf("unexpected contact in event: %+v", evt.ChatItems[0].ChatInfo.Contact)
	}
}
