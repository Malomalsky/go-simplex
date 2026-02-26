package command

import "testing"

func TestRefs(t *testing.T) {
	t.Parallel()

	if got, want := DirectRef(42), "@42"; got != want {
		t.Fatalf("DirectRef: got %q want %q", got, want)
	}
	if got, want := GroupRef(7), "#7"; got != want {
		t.Fatalf("GroupRef: got %q want %q", got, want)
	}
	if got, want := LocalRef(5), "*5"; got != want {
		t.Fatalf("LocalRef: got %q want %q", got, want)
	}
}
