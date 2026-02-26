package spec

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type NetworkUsage string

const (
	NetworkUsageNo          NetworkUsage = "no"
	NetworkUsageInteractive NetworkUsage = "interactive"
	NetworkUsageBackground  NetworkUsage = "background"
	NetworkUsageUnknown     NetworkUsage = "unknown"
)

type CommandParam struct {
	Name string
	Type string
}

type CommandSpec struct {
	Name         string
	Description  string
	NetworkUsage NetworkUsage
	Params       []CommandParam
	Syntax       string
}

type CommandCategory struct {
	Name        string
	Description string
	Commands    []CommandSpec
}

type CommandsDoc struct {
	Categories []CommandCategory
}

func (d CommandsDoc) AllCommands() []CommandSpec {
	var out []CommandSpec
	for _, cat := range d.Categories {
		out = append(out, cat.Commands...)
	}
	return out
}

var (
	networkUsageRe = regexp.MustCompile(`^\*Network usage\*: ([a-zA-Z]+)\.$`)
	paramLineRe    = regexp.MustCompile(`^- ([^:]+): (.+)$`)
)

func ParseCommandsMarkdown(r io.Reader) (CommandsDoc, error) {
	lines, err := readLines(r)
	if err != nil {
		return CommandsDoc{}, err
	}

	var doc CommandsDoc
	for i := 0; i < len(lines); {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "## ") {
			i++
			continue
		}

		category := CommandCategory{
			Name: strings.TrimSpace(strings.TrimPrefix(line, "## ")),
		}
		i++

		start := i
		for i < len(lines) {
			t := strings.TrimSpace(lines[i])
			if strings.HasPrefix(t, "### ") || strings.HasPrefix(t, "## ") {
				break
			}
			i++
		}
		category.Description = normalizeBlock(lines[start:i])

		for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "### ") {
			cmd, next, parseErr := parseCommand(lines, i)
			if parseErr != nil {
				return CommandsDoc{}, parseErr
			}
			category.Commands = append(category.Commands, cmd)
			i = next
			for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
				i++
			}
		}

		doc.Categories = append(doc.Categories, category)
	}

	if len(doc.Categories) == 0 {
		return CommandsDoc{}, fmt.Errorf("no command categories found")
	}
	return doc, nil
}

func parseCommand(lines []string, start int) (CommandSpec, int, error) {
	if start >= len(lines) {
		return CommandSpec{}, start, fmt.Errorf("invalid command start index %d", start)
	}

	header := strings.TrimSpace(lines[start])
	if !strings.HasPrefix(header, "### ") {
		return CommandSpec{}, start, fmt.Errorf("expected command header at line %d", start+1)
	}

	cmd := CommandSpec{Name: strings.TrimSpace(strings.TrimPrefix(header, "### "))}
	i := start + 1

	descStart := i
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if networkUsageRe.MatchString(t) {
			break
		}
		if strings.HasPrefix(t, "### ") || strings.HasPrefix(t, "## ") {
			return CommandSpec{}, i, fmt.Errorf("command %q missing network usage section", cmd.Name)
		}
		i++
	}
	if i >= len(lines) {
		return CommandSpec{}, i, fmt.Errorf("command %q missing network usage section", cmd.Name)
	}
	cmd.Description = normalizeBlock(lines[descStart:i])

	match := networkUsageRe.FindStringSubmatch(strings.TrimSpace(lines[i]))
	if len(match) != 2 {
		return CommandSpec{}, i, fmt.Errorf("cannot parse network usage for command %q", cmd.Name)
	}
	cmd.NetworkUsage = parseNetworkUsage(match[1])
	i++

	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "### ") || strings.HasPrefix(t, "## ") {
			break
		}

		if t == "**Parameters**:" {
			params, next, err := parseParams(lines, i+1)
			if err != nil {
				return CommandSpec{}, i, fmt.Errorf("command %q: %w", cmd.Name, err)
			}
			cmd.Params = params
			i = next
			continue
		}

		if t == "**Syntax**:" {
			syntax, next, err := parseFirstCodeFence(lines, i+1)
			if err != nil {
				return CommandSpec{}, i, fmt.Errorf("command %q: %w", cmd.Name, err)
			}
			cmd.Syntax = syntax
			i = next
			continue
		}

		i++
	}

	if cmd.Syntax == "" {
		return CommandSpec{}, i, fmt.Errorf("command %q missing syntax section", cmd.Name)
	}
	return cmd, i, nil
}

func parseParams(lines []string, start int) ([]CommandParam, int, error) {
	var params []CommandParam
	i := start
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "- ") {
			match := paramLineRe.FindStringSubmatch(t)
			if len(match) != 3 {
				return nil, i, fmt.Errorf("cannot parse parameter line %q", t)
			}
			params = append(params, CommandParam{
				Name: strings.TrimSpace(match[1]),
				Type: strings.TrimSpace(match[2]),
			})
			i++
			continue
		}
		if t == "" {
			i++
			continue
		}
		break
	}
	return params, i, nil
}

func parseFirstCodeFence(lines []string, start int) (string, int, error) {
	i := start
	for i < len(lines) && strings.TrimSpace(lines[i]) != "```" {
		i++
	}
	if i >= len(lines) {
		return "", i, fmt.Errorf("opening code fence not found")
	}
	i++
	bodyStart := i
	for i < len(lines) && strings.TrimSpace(lines[i]) != "```" {
		i++
	}
	if i >= len(lines) {
		return "", i, fmt.Errorf("closing code fence not found")
	}
	body := normalizeBlock(lines[bodyStart:i])
	i++
	return body, i, nil
}

func parseNetworkUsage(s string) NetworkUsage {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(NetworkUsageNo):
		return NetworkUsageNo
	case string(NetworkUsageInteractive):
		return NetworkUsageInteractive
	case string(NetworkUsageBackground):
		return NetworkUsageBackground
	default:
		return NetworkUsageUnknown
	}
}

func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func normalizeBlock(lines []string) string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
