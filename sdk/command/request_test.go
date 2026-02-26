package command

import "testing"

func TestGeneratedCommandStrings(t *testing.T) {
	t.Parallel()

	if got, want := (ShowActiveUser{}).CommandString(), "/user"; got != want {
		t.Fatalf("ShowActiveUser string: got %q want %q", got, want)
	}
	if got, want := (APICreateMyAddress{UserId: 9}).CommandString(), "/_address 9"; got != want {
		t.Fatalf("APICreateMyAddress string: got %q want %q", got, want)
	}
	if got, want := (APIDeleteMyAddress{UserId: 7}).CommandString(), "/_delete_address 7"; got != want {
		t.Fatalf("APIDeleteMyAddress string: got %q want %q", got, want)
	}
}

func TestGeneratedCommandStringsComplex(t *testing.T) {
	t.Parallel()

	ttl := int64(30)
	send := APISendMessages{
		SendRef:     "@42",
		LiveMessage: true,
		Ttl:         &ttl,
		ComposedMessages: []map[string]any{
			{
				"msgContent": map[string]any{
					"type": "text",
					"text": "hi",
				},
				"mentions": map[string]any{},
			},
		},
	}
	if got, want := send.CommandString(), "/_send @42 live=on ttl=30 json [{\"mentions\":{},\"msgContent\":{\"text\":\"hi\",\"type\":\"text\"}}]"; got != want {
		t.Fatalf("APISendMessages string: got %q want %q", got, want)
	}

	storeEncrypted := false
	fileInline := true
	filePath := "/tmp/photo.jpg"
	receive := ReceiveFile{
		FileId:             99,
		UserApprovedRelays: true,
		StoreEncrypted:     &storeEncrypted,
		FileInline:         &fileInline,
		FilePath:           &filePath,
	}
	if got, want := receive.CommandString(), "/freceive 99 approved_relays=on encrypt=off inline=on /tmp/photo.jpg"; got != want {
		t.Fatalf("ReceiveFile string: got %q want %q", got, want)
	}
	if got, want := (ReceiveFile{FileId: 11}).CommandString(), "/freceive 11"; got != want {
		t.Fatalf("ReceiveFile minimal string: got %q want %q", got, want)
	}

	contactID := int64(7)
	search := "support"
	listGroups := APIListGroups{
		UserId:     5,
		ContactId_: &contactID,
		Search:     &search,
	}
	if got, want := listGroups.CommandString(), "/_groups 5 @7 support"; got != want {
		t.Fatalf("APIListGroups string: got %q want %q", got, want)
	}

	viewPwd := "pwd123"
	deleteUser := APIDeleteUser{
		UserId:       5,
		DelSMPQueues: false,
		ViewPwd:      &viewPwd,
	}
	if got, want := deleteUser.CommandString(), "/_delete user 5 del_smp=off \"pwd123\""; got != want {
		t.Fatalf("APIDeleteUser string: got %q want %q", got, want)
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()

	def, ok := Lookup("APISendMessages")
	if !ok {
		t.Fatalf("expected APISendMessages definition")
	}
	if def.Name != "APISendMessages" {
		t.Fatalf("unexpected definition: %+v", def)
	}
	if _, ok := Lookup("DoesNotExist"); ok {
		t.Fatalf("unexpected definition for unknown command")
	}
}
