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

func TestParseRef(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		kind RefKind
		id   int64
	}{
		{in: "@42", kind: RefKindDirect, id: 42},
		{in: "#7", kind: RefKindGroup, id: 7},
		{in: "*5", kind: RefKindLocal, id: 5},
	}

	for _, tc := range cases {
		got, err := ParseRef(tc.in)
		if err != nil {
			t.Fatalf("ParseRef(%q): %v", tc.in, err)
		}
		if got.Kind != tc.kind || got.ID != tc.id {
			t.Fatalf("ParseRef(%q): got kind=%s id=%d", tc.in, got.Kind, got.ID)
		}
	}
}

func TestParseRefInvalid(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"",
		"@",
		"42",
		"@-1",
		"@ 1",
		"#abc",
		"*@1",
	}
	for _, in := range invalid {
		if _, err := ParseRef(in); err == nil {
			t.Fatalf("ParseRef(%q): expected error", in)
		}
	}
}
