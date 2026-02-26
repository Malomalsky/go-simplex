package spec

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

func TestExtractTRefs(t *testing.T) {
	t.Parallel()

	refs := ExtractTRefs("T.User | T.Contact[] | T.User")
	if len(refs) != 2 {
		t.Fatalf("refs: got %d want 2", len(refs))
	}
	if refs[0] != "Contact" || refs[1] != "User" {
		t.Fatalf("unexpected refs: %#v", refs)
	}
}

func TestCollectAndExpandTypeRefs(t *testing.T) {
	t.Parallel()

	respFile, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "responses.ts"))
	if err != nil {
		t.Fatalf("open responses.ts: %v", err)
	}
	defer respFile.Close()
	responses, err := ParseTSInterfaces(respFile)
	if err != nil {
		t.Fatalf("parse responses interfaces: %v", err)
	}

	evtFile, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "events.ts"))
	if err != nil {
		t.Fatalf("open events.ts: %v", err)
	}
	defer evtFile.Close()
	events, err := ParseTSInterfaces(evtFile)
	if err != nil {
		t.Fatalf("parse events interfaces: %v", err)
	}

	typesFile, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "types.ts"))
	if err != nil {
		t.Fatalf("open types.ts: %v", err)
	}
	defer typesFile.Close()
	typesIfaces, err := ParseTopLevelTSInterfaces(typesFile)
	if err != nil {
		t.Fatalf("parse top-level types interfaces: %v", err)
	}

	seed := CollectTRefs(responses)
	seed = append(seed, CollectTRefs(events)...)
	closure := ExpandTypeClosure(typesIfaces, seed)
	if len(closure) == 0 {
		t.Fatalf("expected non-empty closure")
	}

	expect := map[string]bool{
		"GroupInfo":       false,
		"GroupMember":     false,
		"RcvFileTransfer": false,
	}
	for _, name := range closure {
		if _, ok := expect[name]; ok {
			expect[name] = true
		}
	}
	for name, ok := range expect {
		if !ok {
			t.Fatalf("missing type in closure: %s", name)
		}
	}
}

func TestSeedRefsResolvableAgainstTypesSnapshot(t *testing.T) {
	t.Parallel()

	respFile, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "responses.ts"))
	if err != nil {
		t.Fatalf("open responses.ts: %v", err)
	}
	defer respFile.Close()
	responses, err := ParseTSInterfaces(respFile)
	if err != nil {
		t.Fatalf("parse responses interfaces: %v", err)
	}

	evtFile, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "events.ts"))
	if err != nil {
		t.Fatalf("open events.ts: %v", err)
	}
	defer evtFile.Close()
	events, err := ParseTSInterfaces(evtFile)
	if err != nil {
		t.Fatalf("parse events interfaces: %v", err)
	}

	typesFile, err := os.Open(filepath.Join("..", "..", "spec", "upstream", "types.ts"))
	if err != nil {
		t.Fatalf("open types.ts: %v", err)
	}
	defer typesFile.Close()
	typeIfaces, err := ParseTopLevelTSInterfaces(typesFile)
	if err != nil {
		t.Fatalf("parse top-level types interfaces: %v", err)
	}

	known := BuiltinTypeNameSet()
	for _, iface := range typeIfaces {
		known[iface.Name] = struct{}{}
	}
	for _, name := range parseTopLevelTypeOrEnumNames(t, filepath.Join("..", "..", "spec", "upstream", "types.ts")) {
		known[name] = struct{}{}
	}

	seed := append(CollectTRefs(responses), CollectTRefs(events)...)
	missing := map[string]struct{}{}
	for _, ref := range seed {
		if _, ok := known[ref]; !ok {
			missing[ref] = struct{}{}
		}
	}

	if len(missing) > 0 {
		out := make([]string, 0, len(missing))
		for name := range missing {
			out = append(out, name)
		}
		sort.Strings(out)
		t.Fatalf("unresolved T.* refs from response/event snapshots: %v", out)
	}
}

func parseTopLevelTypeOrEnumNames(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	re := regexp.MustCompile(`^export (?:type|enum) ([A-Za-z0-9_]+)\b`)
	seen := map[string]struct{}{}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		m := re.FindStringSubmatch(sc.Text())
		if len(m) == 2 {
			seen[m[1]] = struct{}{}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}

	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
