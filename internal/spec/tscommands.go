package spec

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type TSCommand struct {
	Name         string
	Fields       []TSField
	ExprJS       string
	Category     string
	Responses    []TSCommandResponse
	ResponseTags []string
}

type TSCommandResponse struct {
	Name string
	Tag  string
}

var (
	cmdIfaceStartRe = regexp.MustCompile(`^\s*export interface ([A-Za-z0-9_]+) \{$`)
	cmdNsStartRe    = regexp.MustCompile(`^\s*export namespace ([A-Za-z0-9_]+) \{$`)
	cmdRespTypeRe   = regexp.MustCompile(`^\s*export type Response = (.+?);?\s*$`)
	cmdReturnRe     = regexp.MustCompile(`^\s*return (.+?);?\s*$`)
)

func ParseTSCommands(r io.Reader, catalog CommandsDoc, responseTypes []TaggedType) ([]TSCommand, error) {
	lines, err := readCmdLines(r)
	if err != nil {
		return nil, err
	}

	responseTagByName := make(map[string]string, len(responseTypes))
	for _, rt := range responseTypes {
		responseTagByName[rt.Name] = rt.Tag
	}

	ifaceMap := make(map[string][]TSField, 64)
	exprMap := make(map[string]string, 64)
	respMap := make(map[string][]TSCommandResponse, 64)

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
			var respTypes []TSCommandResponse
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "}" {
					i = j
					break
				}

				if rm := cmdRespTypeRe.FindStringSubmatch(next); len(rm) == 2 {
					types, parseErr := parseNamespaceResponseTypes(rm[1], responseTagByName)
					if parseErr != nil {
						return nil, fmt.Errorf("parse command namespace response %s: %w", name, parseErr)
					}
					respTypes = types
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
			if len(respTypes) == 0 {
				return nil, fmt.Errorf("command namespace %s has no response type union", name)
			}
			respMap[name] = respTypes
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
		respTypes, ok := respMap[c.Name]
		if !ok {
			return nil, fmt.Errorf("command response union not found in commands.ts: %s", c.Name)
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
			Name:         c.Name,
			Fields:       fields,
			ExprJS:       expr,
			Category:     category,
			Responses:    append([]TSCommandResponse(nil), respTypes...),
			ResponseTags: extractResponseTags(respTypes),
		})
	}

	return out, nil
}

func parseNamespaceResponseTypes(responseExpr string, responseTagByName map[string]string) ([]TSCommandResponse, error) {
	parts := strings.Split(responseExpr, "|")
	out := make([]TSCommandResponse, 0, len(parts))
	for _, p := range parts {
		part := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(p), ";"))
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "CR.") {
			part = strings.TrimPrefix(part, "CR.")
		}
		tag, ok := responseTagByName[part]
		if !ok {
			return nil, fmt.Errorf("unknown response type %q", part)
		}
		out = append(out, TSCommandResponse{
			Name: part,
			Tag:  tag,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty response type union")
	}
	return out, nil
}

func extractResponseTags(types []TSCommandResponse) []string {
	out := make([]string, 0, len(types))
	for _, rt := range types {
		out = append(out, rt.Tag)
	}
	return out
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
