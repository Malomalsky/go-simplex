package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewWebSocket(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req struct {
			CorrID string `json:"corrId"`
			Cmd    string `json:"cmd"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"corrId":"`+req.CorrID+`","resp":{"type":"cmdOk"}}`))
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c, err := NewWebSocket(ctx, wsURL)
	if err != nil {
		t.Fatalf("new websocket client: %v", err)
	}
	defer c.Close(context.Background())

	msg, err := c.SendRaw(ctx, "/user")
	if err != nil {
		t.Fatalf("send raw: %v", err)
	}
	if msg.Resp.Type != "cmdOk" {
		t.Fatalf("unexpected response type: %s", msg.Resp.Type)
	}
}
