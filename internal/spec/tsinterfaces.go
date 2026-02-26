package spec

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type TSField struct {
	Name     string
	TypeExpr string
	Optional bool
	Comment  string
}

type TSInterface struct {
	Name   string
	Fields []TSField
}

var (
	tsIfaceStartRe = regexp.MustCompile(`^\s*export interface ([A-Za-z0-9_]+)(?: extends Interface)? \{$`)
	tsFieldRe      = regexp.MustCompile(`^\s*([A-Za-z0-9_]+)(\?)?:\s*([^/]+?)(?:\s*//\s*(.+))?\s*$`)
)

func ParseTSInterfaces(r io.Reader) ([]TSInterface, error) {
	lines, err := readTSLines(r)
	if err != nil {
		return nil, err
	}

	out := make([]TSInterface, 0, 64)
	for i := 0; i < len(lines); i++ {
		start := tsIfaceStartRe.FindStringSubmatch(lines[i])
		if len(start) != 2 {
			continue
		}

		iface := TSInterface{Name: start[1]}
		for j := i + 1; j < len(lines); j++ {
			line := strings.TrimSpace(lines[j])
			if line == "}" {
				i = j
				break
			}
			if line == "" {
				continue
			}

			fieldMatch := tsFieldRe.FindStringSubmatch(lines[j])
			if len(fieldMatch) != 5 {
				return nil, fmt.Errorf("cannot parse TS field line: %q", lines[j])
			}

			iface.Fields = append(iface.Fields, TSField{
				Name:     fieldMatch[1],
				Optional: fieldMatch[2] == "?",
				TypeExpr: strings.TrimSpace(fieldMatch[3]),
				Comment:  strings.TrimSpace(fieldMatch[4]),
			})
		}

		if len(iface.Fields) == 0 {
			return nil, fmt.Errorf("interface %s has no fields", iface.Name)
		}
		out = append(out, iface)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no interfaces parsed")
	}
	return out, nil
}

func readTSLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	lines := make([]string, 0, 512)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
