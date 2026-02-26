package protocol

import (
	"encoding/json"
	"testing"
)

func TestEncodeRequest(t *testing.T) {
	t.Parallel()

	payload, err := EncodeRequest(CommandRequest{
		CorrID: "1",
		Cmd:    "/user",
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}

	var req CommandRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	if req.CorrID != "1" || req.Cmd != "/user" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestDecodeMessage(t *testing.T) {
	t.Parallel()

	msg, err := DecodeMessage([]byte(`{"corrId":"12","resp":{"type":"activeUser","user":{"userId":1}}}`))
	if err != nil {
		t.Fatalf("decode message: %v", err)
	}
	if !msg.IsResponse() || msg.IsEvent() {
		t.Fatalf("expected correlated response")
	}
	if got, want := msg.Resp.Type, "activeUser"; got != want {
		t.Fatalf("resp type: got %q want %q", got, want)
	}

	var decoded struct {
		Type string `json:"type"`
		User struct {
			UserID int64 `json:"userId"`
		} `json:"user"`
	}
	if err := msg.Resp.Decode(&decoded); err != nil {
		t.Fatalf("decode raw response: %v", err)
	}
	if got, want := decoded.User.UserID, int64(1); got != want {
		t.Fatalf("decoded user id: got %d want %d", got, want)
	}
}

func TestDecodeMessageRequiresType(t *testing.T) {
	t.Parallel()

	_, err := DecodeMessage([]byte(`{"resp":{"user":{"userId":1}}}`))
	if err == nil {
		t.Fatalf("expected error for missing response type")
	}
}
