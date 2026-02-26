package spec

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type TaggedType struct {
	Name string
	Tag  string
}

var (
	ifaceStartRe = regexp.MustCompile(`^\s*export interface ([A-Za-z0-9_]+) extends Interface \{$`)
	typeLineRe   = regexp.MustCompile(`^\s*type:\s*"([^"]+)"`)
)

func ParseTaggedInterfaces(r io.Reader) ([]TaggedType, error) {
	sc := bufio.NewScanner(r)
	lines := make([]string, 0, 512)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	out := make([]TaggedType, 0, 64)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		start := ifaceStartRe.FindStringSubmatch(line)
		if len(start) != 2 {
			continue
		}
		name := start[1]
		tag := ""
		for j := i + 1; j < len(lines); j++ {
			next := strings.TrimSpace(lines[j])
			if next == "}" {
				i = j
				break
			}
			match := typeLineRe.FindStringSubmatch(next)
			if len(match) == 2 {
				tag = match[1]
			}
		}
		if tag == "" {
			return nil, fmt.Errorf("interface %s has no type tag", name)
		}
		out = append(out, TaggedType{Name: name, Tag: tag})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no tagged interfaces found")
	}
	return out, nil
}
