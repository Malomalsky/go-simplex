package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Malomalsky/go-simplex/internal/spec"
)

func main() {
	var (
		inputPath  string
		outputPath string
		pkgName    string
		eventsPath string
		respPath   string
		tagsOut    string
		recordsOut string
		cmdTSPath  string
		reqOut     string
	)

	flag.StringVar(&inputPath, "input", "spec/upstream/COMMANDS.md", "path to COMMANDS.md")
	flag.StringVar(&outputPath, "output", "sdk/command/generated_catalog.go", "generated go file path")
	flag.StringVar(&pkgName, "package", "command", "package name for generated file")
	flag.StringVar(&eventsPath, "events", "spec/upstream/events.ts", "path to upstream events.ts")
	flag.StringVar(&respPath, "responses", "spec/upstream/responses.ts", "path to upstream responses.ts")
	flag.StringVar(&tagsOut, "out-tags", "sdk/types/generated_tags.go", "generated event/response tag constants file")
	flag.StringVar(&recordsOut, "out-records", "sdk/types/generated_records.go", "generated event/response records file")
	flag.StringVar(&cmdTSPath, "commands-ts", "spec/upstream/commands.ts", "path to upstream commands.ts")
	flag.StringVar(&reqOut, "out-requests", "sdk/command/generated_requests.go", "generated command request structs file")
	flag.Parse()

	if err := run(inputPath, outputPath, pkgName, eventsPath, respPath, tagsOut, recordsOut, cmdTSPath, reqOut); err != nil {
		fmt.Fprintf(os.Stderr, "simplexgen: %v\n", err)
		os.Exit(1)
	}
}

func run(inputPath, outputPath, pkgName, eventsPath, respPath, tagsOut, recordsOut, cmdTSPath, reqOut string) error {
	in, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer in.Close()

	doc, err := spec.ParseCommandsMarkdown(in)
	if err != nil {
		return fmt.Errorf("parse commands markdown: %w", err)
	}

	code, err := spec.RenderCatalogGo(doc, pkgName)
	if err != nil {
		return fmt.Errorf("render go catalog: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, code, 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	responses, err := parseTaggedTS(respPath)
	if err != nil {
		return fmt.Errorf("parse responses tags: %w", err)
	}
	events, err := parseTaggedTS(eventsPath)
	if err != nil {
		return fmt.Errorf("parse events tags: %w", err)
	}

	tags, err := spec.RenderTypesTagsGo("types", responses, events)
	if err != nil {
		return fmt.Errorf("render tags file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(tagsOut), 0o755); err != nil {
		return fmt.Errorf("create tags output directory: %w", err)
	}
	if err := os.WriteFile(tagsOut, tags, 0o644); err != nil {
		return fmt.Errorf("write tags output file: %w", err)
	}

	responseIfaces, err := parseTSInterfaces(respPath)
	if err != nil {
		return fmt.Errorf("parse response interfaces: %w", err)
	}
	eventIfaces, err := parseTSInterfaces(eventsPath)
	if err != nil {
		return fmt.Errorf("parse event interfaces: %w", err)
	}
	records, err := spec.RenderTypesRecordsGo("types", responseIfaces, eventIfaces)
	if err != nil {
		return fmt.Errorf("render records file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(recordsOut), 0o755); err != nil {
		return fmt.Errorf("create records output directory: %w", err)
	}
	if err := os.WriteFile(recordsOut, records, 0o644); err != nil {
		return fmt.Errorf("write records output file: %w", err)
	}

	cmdTSFile, err := os.Open(cmdTSPath)
	if err != nil {
		return fmt.Errorf("open commands.ts: %w", err)
	}
	defer cmdTSFile.Close()

	tsCommands, err := spec.ParseTSCommands(cmdTSFile, doc)
	if err != nil {
		return fmt.Errorf("parse ts commands: %w", err)
	}
	reqCode, err := spec.RenderCommandRequestsGo(pkgName, tsCommands)
	if err != nil {
		return fmt.Errorf("render command requests: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(reqOut), 0o755); err != nil {
		return fmt.Errorf("create requests output directory: %w", err)
	}
	if err := os.WriteFile(reqOut, reqCode, 0o644); err != nil {
		return fmt.Errorf("write requests output file: %w", err)
	}
	return nil
}

func parseTaggedTS(path string) ([]spec.TaggedType, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return spec.ParseTaggedInterfaces(f)
}

func parseTSInterfaces(path string) ([]spec.TSInterface, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return spec.ParseTSInterfaces(f)
}
