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

