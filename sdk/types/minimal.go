package types

import "encoding/json"

type Profile struct {
	DisplayName string `json:"displayName,omitempty"`
	FullName    string `json:"fullName,omitempty"`
}

type User struct {
	UserID  int64   `json:"userId"`
	Profile Profile `json:"profile"`
}

type CreatedConnLink struct {
	ConnFullLink  string `json:"connFullLink,omitempty"`
	ConnShortLink string `json:"connShortLink,omitempty"`
}

func (l CreatedConnLink) PreferredLink() string {
	if l.ConnShortLink != "" {
		return l.ConnShortLink
	}
	return l.ConnFullLink
}

type UserContactLink struct {
	ConnLinkContact CreatedConnLink `json:"connLinkContact"`
}

type Contact struct {
	ContactID int64   `json:"contactId"`
	Profile   Profile `json:"profile"`
}

type ChatError struct {
	Type       string          `json:"type"`
	ErrorType  json.RawMessage `json:"errorType,omitempty"`
	StoreError json.RawMessage `json:"storeError,omitempty"`
}

type ChatInfo struct {
	Type    string   `json:"type"`
	Contact *Contact `json:"contact,omitempty"`
}

type MsgContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ChatContent struct {
	Type       string      `json:"type"`
	MsgContent *MsgContent `json:"msgContent,omitempty"`
}

type AChatItem struct {
	ChatInfo ChatInfo `json:"chatInfo"`
	ChatItem ChatItem `json:"chatItem"`
}

type ChatItem struct {
	Content ChatContent `json:"content"`
}
