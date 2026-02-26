package spec

import (
	"regexp"
	"sort"
	"strings"
)

var tRefRe = regexp.MustCompile(`\bT\.([A-Za-z0-9_]+)\b`)

var builtinTypes = map[string]struct{}{
	"Profile":         {},
	"User":            {},
	"CreatedConnLink": {},
	"UserContactLink": {},
	"Contact":         {},
	"ChatError":       {},
	"ChatInfo":        {},
	"MsgContent":      {},
	"ChatContent":     {},
	"AChatItem":       {},
	"ChatItem":        {},
}

func BuiltinTypeNameSet() map[string]struct{} {
	out := make(map[string]struct{}, len(builtinTypes))
	for name := range builtinTypes {
		out[name] = struct{}{}
	}
	return out
}

func ExtractTRefs(typeExpr string) []string {
	matches := tRefRe.FindAllStringSubmatch(typeExpr, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) != 2 {
			continue
		}
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func CollectTRefs(ifaces []TSInterface) []string {
	seen := map[string]struct{}{}
	for _, iface := range ifaces {
		for _, field := range iface.Fields {
			for _, ref := range ExtractTRefs(field.TypeExpr) {
				seen[ref] = struct{}{}
			}
		}
	}
	return sortedKeys(seen)
}

func ExpandTypeClosure(typeIfaces []TSInterface, seedRefs []string) []string {
	index := make(map[string]TSInterface, len(typeIfaces))
	for _, iface := range typeIfaces {
		index[iface.Name] = iface
	}

	seen := map[string]struct{}{}
	queue := make([]string, 0, len(seedRefs))
	for _, name := range seedRefs {
		if _, ok := index[name]; !ok {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		queue = append(queue, name)
	}

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		iface := index[name]
		for _, field := range iface.Fields {
			for _, ref := range ExtractTRefs(field.TypeExpr) {
				if _, ok := index[ref]; !ok {
					continue
				}
				if _, ok := seen[ref]; ok {
					continue
				}
				seen[ref] = struct{}{}
				queue = append(queue, ref)
			}
		}
	}
	return sortedKeys(seen)
}

func FilterTypeNames(names []string, exclude map[string]struct{}) []string {
	if len(names) == 0 {
		return nil
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if _, drop := exclude[name]; drop {
			continue
		}
		out = append(out, name)
	}
	return out
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
