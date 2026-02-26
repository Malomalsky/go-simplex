package spec

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type TSCommand struct {
	Name     string
	Fields   []TSField
	ExprJS   string
	Category string
}

var (
	cmdIfaceStartRe = regexp.MustCompile(`^\s*export interface ([A-Za-z0-9_]+) \{$`)
	cmdNsStartRe    = regexp.MustCompile(`^\s*export namespace ([A-Za-z0-9_]+) \{$`)
	cmdReturnRe     = regexp.MustCompile(`^\s*return (.+?);?\s*$`)
)

func ParseTSCommands(r io.Reader, catalog CommandsDoc) ([]TSCommand, error) {
	lines, err := readCmdLines(r)
	if err != nil {
		return nil, err
	}

	ifaceMap := make(map[string][]TSField, 64)
	exprMap := make(map[string]string, 64)

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if m := cmdIfaceStartRe.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			fields := make([]TSField, 0, 8)
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "}" {
					i = j
					break
				}
				if next == "" {
					continue
				}
				fm := tsFieldRe.FindStringSubmatch(lines[j])
				if len(fm) != 5 {
					return nil, fmt.Errorf("cannot parse command interface field: %q", lines[j])
				}
				fields = append(fields, TSField{
					Name:     fm[1],
					Optional: fm[2] == "?",
					TypeExpr: strings.TrimSpace(fm[3]),
					Comment:  strings.TrimSpace(fm[4]),
				})
			}
			ifaceMap[name] = fields
			continue
		}

		if m := cmdNsStartRe.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			found := false
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "}" {
					i = j
					break
				}
				rm := cmdReturnRe.FindStringSubmatch(next)
				if len(rm) == 2 {
					exprMap[name] = strings.TrimSpace(rm[1])
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("command namespace %s has no return expression", name)
			}
		}
	}

	// Preserve documented order using parsed catalog.
	out := make([]TSCommand, 0, len(catalog.AllCommands()))
	for _, c := range catalog.AllCommands() {
		fields, ok := ifaceMap[c.Name]
		if !ok {
			return nil, fmt.Errorf("command interface not found in commands.ts: %s", c.Name)
		}
		expr, ok := exprMap[c.Name]
		if !ok {
			return nil, fmt.Errorf("command expression not found in commands.ts: %s", c.Name)
		}

		category := ""
		for _, cat := range catalog.Categories {
			for _, cmd := range cat.Commands {
				if cmd.Name == c.Name {
					category = cat.Name
					break
				}
			}
			if category != "" {
				break
			}
		}

		out = append(out, TSCommand{
			Name:     c.Name,
			Fields:   fields,
			ExprJS:   expr,
			Category: category,
		})
	}

	return out, nil
}

func readCmdLines(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	lines := make([]string, 0, 1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
