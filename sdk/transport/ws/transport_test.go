package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestTransportReadWrite(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.TextMessage {
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, payload)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tr, err := Dial(ctx, wsURL)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer tr.Close()

	if err := tr.Write(ctx, []byte(`{"corrId":"1","cmd":"/user"}`)); err != nil {
		t.Fatalf("write: %v", err)
	}
	payload, err := tr.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got, want := string(payload), `{"corrId":"1","cmd":"/user"}`; got != want {
		t.Fatalf("payload mismatch: got %q want %q", got, want)
	}
}

func TestTransportReadRejectsBinary(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_ = conn.WriteMessage(websocket.BinaryMessage, []byte{0x01, 0x02})
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tr, err := Dial(ctx, wsURL)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer tr.Close()

	_, err = tr.Read(ctx)
	if err == nil {
		t.Fatalf("expected error for binary frame")
	}
}

func TestDialRejectsUnsupportedScheme(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := Dial(ctx, "http://localhost:5225")
	if err == nil {
		t.Fatalf("expected unsupported scheme error")
	}
}

func TestDialRequireWSSRejectsWS(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := Dial(ctx, wsURL, WithRequireWSS(true))
	if err == nil {
		t.Fatalf("expected require-wss rejection")
	}
}
