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
	)

	flag.StringVar(&inputPath, "input", "spec/upstream/COMMANDS.md", "path to COMMANDS.md")
	flag.StringVar(&outputPath, "output", "sdk/command/generated_catalog.go", "generated go file path")
	flag.StringVar(&pkgName, "package", "command", "package name for generated file")
	flag.StringVar(&eventsPath, "events", "spec/upstream/events.ts", "path to upstream events.ts")
	flag.StringVar(&respPath, "responses", "spec/upstream/responses.ts", "path to upstream responses.ts")
	flag.StringVar(&tagsOut, "out-tags", "sdk/types/generated_tags.go", "generated event/response tag constants file")
	flag.Parse()

	if err := run(inputPath, outputPath, pkgName, eventsPath, respPath, tagsOut); err != nil {
		fmt.Fprintf(os.Stderr, "simplexgen: %v\n", err)
		os.Exit(1)
	}
}

func run(inputPath, outputPath, pkgName, eventsPath, respPath, tagsOut string) error {
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
