package command

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Request interface {
	CommandString() string
}

type Raw string

func (r Raw) CommandString() string {
	return string(r)
}

func Lookup(name string) (Definition, bool) {
	for _, def := range GeneratedCatalog {
		if def.Name == name {
			return def, true
		}
	}
	return Definition{}, false
}

type ShowActiveUser struct{}

func (ShowActiveUser) CommandString() string { return "/user" }

type APIShowMyAddress struct {
	UserID int64
}

func (c APIShowMyAddress) CommandString() string {
	return "/_show_address " + strconv.FormatInt(c.UserID, 10)
}

type APICreateMyAddress struct {
	UserID int64
}

func (c APICreateMyAddress) CommandString() string {
	return "/_address " + strconv.FormatInt(c.UserID, 10)
}

type APIDeleteMyAddress struct {
	UserID int64
}

func (c APIDeleteMyAddress) CommandString() string {
	return "/_delete_address " + strconv.FormatInt(c.UserID, 10)
}

type APISetAddressSettings struct {
	UserID   int64
	Settings any
}

func (c APISetAddressSettings) CommandString() string {
	return "/_address_settings " + strconv.FormatInt(c.UserID, 10) + " " + mustJSON(c.Settings)
}

type APISendMessages struct {
	SendRef          string
	LiveMessage      bool
	TTL              *int
	ComposedMessages any
}

func (c APISendMessages) CommandString() string {
	var b strings.Builder
	b.WriteString("/_send ")
	b.WriteString(c.SendRef)
	if c.LiveMessage {
		b.WriteString(" live=on")
	}
	if c.TTL != nil {
		b.WriteString(" ttl=")
		b.WriteString(strconv.Itoa(*c.TTL))
	}
	b.WriteString(" json ")
	b.WriteString(mustJSON(c.ComposedMessages))
	return b.String()
}

func mustJSON(v any) string {
	if v == nil {
		return "null"
	}
	payload, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("json marshal failed: %v", err))
	}
	return string(payload)
}
