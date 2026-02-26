package spec

import (
	"strings"
	"testing"
)

func TestRenderClientSendersGo(t *testing.T) {
	t.Parallel()

	cmds := []TSCommand{
		{
			Name: "ShowActiveUser",
			Responses: []TSCommandResponse{
				{Name: "ActiveUser", Tag: "activeUser"},
				{Name: "ChatCmdError", Tag: "chatCmdError"},
			},
		},
	}

	src, err := RenderClientSendersGo("client", cmds)
	if err != nil {
		t.Fatalf("render client senders: %v", err)
	}
	code := string(src)
	if !strings.Contains(code, "type ShowActiveUserResult struct") {
		t.Fatalf("missing result type")
	}
	if !strings.Contains(code, "func (c *Client) SendShowActiveUser(") {
		t.Fatalf("missing sender method")
	}
	if !strings.Contains(code, "case types.ResponseTypeActiveUser:") {
		t.Fatalf("missing response switch case")
	}
	if !strings.Contains(code, "default:\n\t\treturn out, nil") {
		t.Fatalf("missing tolerant default branch")
	}
}
