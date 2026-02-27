package command

import (
	"fmt"
	"strconv"
	"strings"
)

type RefKind string

const (
	RefKindDirect RefKind = "direct"
	RefKindGroup  RefKind = "group"
	RefKindLocal  RefKind = "local"
)

type ParsedRef struct {
	Raw  string
	Kind RefKind
	ID   int64
}

func DirectRef(contactID int64) string {
	return "@" + strconv.FormatInt(contactID, 10)
}

func GroupRef(groupID int64) string {
	return "#" + strconv.FormatInt(groupID, 10)
}

func LocalRef(folderID int64) string {
	return "*" + strconv.FormatInt(folderID, 10)
}

func ParseRef(ref string) (ParsedRef, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ParsedRef{}, fmt.Errorf("ref is empty")
	}
	if len(ref) < 2 {
		return ParsedRef{}, fmt.Errorf("ref is too short: %q", ref)
	}

	var kind RefKind
	switch ref[0] {
	case '@':
		kind = RefKindDirect
	case '#':
		kind = RefKindGroup
	case '*':
		kind = RefKindLocal
	default:
		return ParsedRef{}, fmt.Errorf("unsupported ref prefix %q", ref[0])
	}

	idPart := ref[1:]
	if idPart == "" {
		return ParsedRef{}, fmt.Errorf("ref id is empty")
	}
	for _, r := range idPart {
		if r < '0' || r > '9' {
			return ParsedRef{}, fmt.Errorf("ref id contains non-digit %q", r)
		}
	}

	id, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		return ParsedRef{}, fmt.Errorf("parse ref id: %w", err)
	}

	return ParsedRef{
		Raw:  ref,
		Kind: kind,
		ID:   id,
	}, nil
}

func ValidateRef(ref string) error {
	_, err := ParseRef(ref)
	return err
}
