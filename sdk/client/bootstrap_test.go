package client

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestBootstrapBot(t *testing.T) {
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
		// 1) /user
		raw1 := <-transport.writeCh
		var req1 struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(raw1, &req1)
		transport.readCh <- []byte(`{"corrId":"` + req1.CorrID + `","resp":{"type":"activeUser","user":{"userId":10,"profile":{"displayName":"bot"}}}}`)

		// 2) /_show_address
		raw2 := <-transport.writeCh
		var req2 struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(raw2, &req2)
		transport.readCh <- []byte(`{"corrId":"` + req2.CorrID + `","resp":{"type":"chatCmdError","chatError":{"type":"errorStore","storeError":{"type":"userContactLinkNotFound"}}}}`)

		// 3) /_address
		raw3 := <-transport.writeCh
		var req3 struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(raw3, &req3)
		transport.readCh <- []byte(`{"corrId":"` + req3.CorrID + `","resp":{"type":"userContactLinkCreated","user":{"userId":10},"connLinkContact":{"connShortLink":"smp://short"}}}`)

		// 4) /_address_settings
		raw4 := <-transport.writeCh
		var req4 struct {
			CorrID string `json:"corrId"`
		}
		_ = json.Unmarshal(raw4, &req4)
		transport.readCh <- []byte(`{"corrId":"` + req4.CorrID + `","resp":{"type":"userContactLinkUpdated","user":{"userId":10},"contactLink":{"connLinkContact":{"connShortLink":"smp://short"}}}}`)
	}()

	result, err := c.BootstrapBot(ctx)
	if err != nil {
		t.Fatalf("BootstrapBot: %v", err)
	}
	if result.User == nil || result.User.UserID != 10 {
		t.Fatalf("unexpected user: %+v", result.User)
	}
	if result.Address != "smp://short" {
		t.Fatalf("unexpected address: %q", result.Address)
	}
}
