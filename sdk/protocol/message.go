package protocol

import (
	"encoding/json"
	"fmt"
)

type CommandRequest struct {
	CorrID string `json:"corrId"`
	Cmd    string `json:"cmd"`
}

type Message struct {
	CorrID string      `json:"corrId,omitempty"`
	Resp   RawResponse `json:"resp"`
}

func (m Message) IsEvent() bool {
	return m.CorrID == ""
}

func (m Message) IsResponse() bool {
	return m.CorrID != ""
}

type RawResponse struct {
	Type string
	Raw  json.RawMessage
}

func (r *RawResponse) UnmarshalJSON(data []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("unmarshal response probe: %w", err)
	}
	if probe.Type == "" {
		return fmt.Errorf("response type is empty")
	}
	r.Type = probe.Type
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

func EncodeRequest(req CommandRequest) ([]byte, error) {
	if req.CorrID == "" {
		return nil, fmt.Errorf("corrId is required")
	}
	if req.Cmd == "" {
		return nil, fmt.Errorf("cmd is required")
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	return payload, nil
}

func DecodeMessage(payload []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return Message{}, fmt.Errorf("decode message: %w", err)
	}
	return msg, nil
}

