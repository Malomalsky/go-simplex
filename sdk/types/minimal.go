package types

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

type ActiveUserResp struct {
	Type string `json:"type"`
	User User   `json:"user"`
}

type UserContactLinkResp struct {
	Type        string          `json:"type"`
	User        User            `json:"user"`
	ContactLink UserContactLink `json:"contactLink"`
}

type UserContactLinkCreatedResp struct {
	Type            string          `json:"type"`
	User            User            `json:"user"`
	ConnLinkContact CreatedConnLink `json:"connLinkContact"`
}

type UserContactLinkUpdatedResp struct {
	Type        string          `json:"type"`
	User        User            `json:"user"`
	ContactLink UserContactLink `json:"contactLink"`
}

type ChatCmdErrorResp struct {
	Type      string    `json:"type"`
	ChatError ChatError `json:"chatError"`
}

type ChatError struct {
	Type       string      `json:"type"`
	ErrorType  interface{} `json:"errorType,omitempty"`
	StoreError interface{} `json:"storeError,omitempty"`
}

type NewChatItemsResp struct {
	Type      string      `json:"type"`
	ChatItems []AChatItem `json:"chatItems"`
}

type AChatItem struct {
	ChatInfo interface{} `json:"chatInfo"`
	ChatItem ChatItem    `json:"chatItem"`
}

type ChatItem struct {
	Content interface{} `json:"content"`
}
